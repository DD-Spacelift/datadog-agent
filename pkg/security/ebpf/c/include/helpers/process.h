#ifndef _HELPERS_PROCESS_H_
#define _HELPERS_PROCESS_H_

#include "constants/custom.h"
#include "constants/enums.h"
#include "constants/offsets/process.h"
#include "maps.h"
#include "events_definition.h"

#include "container.h"

static __attribute__((always_inline)) void send_signal(u32 pid) {
    if (is_send_signal_available()) {
        u32 *sig = bpf_map_lookup_elem(&kill_list, &pid);
        if ((sig != NULL && *sig != 0)) {
#if defined(DEBUG_SEND_SIGNAL)
            bpf_printk("Sending signal %d to pid %d\n", *sig, pid);
#endif
            bpf_send_signal(*sig);
            bpf_map_delete_elem(&kill_list, &pid);
        }
    }
}

static __attribute__((always_inline)) u32 copy_tty_name(const char src[TTY_NAME_LEN], char dst[TTY_NAME_LEN]) {
    if (src[0] == 0) {
        return 0;
    }

    bpf_probe_read(dst, TTY_NAME_LEN, (void *)src);
    return TTY_NAME_LEN;
}

void __attribute__((always_inline)) copy_proc_entry(struct process_entry_t *src, struct process_entry_t *dst) {
    dst->executable = src->executable;
    dst->exec_timestamp = src->exec_timestamp;
    copy_tty_name(src->tty_name, dst->tty_name);
    bpf_probe_read(dst->comm, TASK_COMM_LEN, src->comm);
}

void __attribute__((always_inline)) copy_proc_cache(struct proc_cache_t *src, struct proc_cache_t *dst) {
    copy_container_id(src->container.container_id, dst->container.container_id);
    dst->container.cgroup_context.cgroup_flags = src->container.cgroup_context.cgroup_flags;
    copy_proc_entry(&src->entry, &dst->entry);
}

void __attribute__((always_inline)) copy_pid_cache_except_exit_ts(struct pid_cache_t *src, struct pid_cache_t *dst) {
    dst->cookie = src->cookie;
    dst->user_session_id = src->user_session_id;
    dst->ppid = src->ppid;
    dst->fork_timestamp = src->fork_timestamp;
    dst->credentials = src->credentials;
}

struct proc_cache_t __attribute__((always_inline)) * get_proc_from_cookie(u64 cookie) {
    if (!cookie) {
        return NULL;
    }

    return bpf_map_lookup_elem(&proc_cache, &cookie);
}

struct pid_cache_t *__attribute__((always_inline)) get_pid_cache(u32 tgid) {
    return (struct pid_cache_t *)bpf_map_lookup_elem(&pid_cache, &tgid);
}

struct proc_cache_t *__attribute__((always_inline)) get_proc_cache(u32 tgid) {
    struct pid_cache_t *pid_entry = get_pid_cache(tgid);
    if (!pid_entry) {
        return NULL;
    }

    // Select the cache entry
    return get_proc_from_cookie(pid_entry->cookie);
}

static struct proc_cache_t *__attribute__((always_inline)) fill_process_context_with_pid_tgid(struct process_context_t *data, u64 pid_tgid) {
    u32 tgid = pid_tgid >> 32;

    // https://github.com/iovisor/bcc/blob/master/docs/reference_guide.md#4-bpf_get_current_pid_tgid
    data->pid = tgid;
    data->tid = pid_tgid;

    u32 tid = data->tid; // This looks unnecessary but it actually is to address this issue https://github.com/iovisor/bcc/issues/347 in at least Ubuntu 4.15.
    u32 *netns = bpf_map_lookup_elem(&netns_cache, &tid);
    if (netns != NULL) {
        data->netns = *netns;
    }

    u32 pid = data->pid;
    // consider kworker a pid which is ignored
    u32 *is_ignored = bpf_map_lookup_elem(&pid_ignored, &pid);
    if (is_ignored) {
        data->is_kworker = 1;
    }

    struct proc_cache_t *pc = get_proc_cache(tgid);
    if (pc) {
        data->inode = pc->entry.executable.path_key.ino;
    }

    return pc;
}

static struct proc_cache_t *__attribute__((always_inline)) fill_process_context(struct process_context_t *data) {
    u64 pid_tgid = bpf_get_current_pid_tgid();
    return fill_process_context_with_pid_tgid(data, pid_tgid);
}

void __attribute__((always_inline)) fill_args_envs(struct process_event_t *event, struct syscall_cache_t *syscall) {
    event->args_id = syscall->exec.args.id;
    event->args_truncated = syscall->exec.args.truncated;
    event->envs_id = syscall->exec.envs.id;
    event->envs_truncated = syscall->exec.envs.truncated;
}

u32 __attribute__((always_inline)) get_root_nr_from_pid_struct(struct pid *pid) {
    // read the root pid namespace nr from &pid->numbers[0].nr
    u32 root_nr = 0;
    bpf_probe_read(&root_nr, sizeof(root_nr), (void *)pid + get_pid_numbers_offset());
    return root_nr;
}

u32 __attribute__((always_inline)) get_root_nr_from_task_struct(struct task_struct *task) {
    struct pid *pid = NULL;
    bpf_probe_read(&pid, sizeof(pid), (void *)task + get_task_struct_pid_offset());
    return get_root_nr_from_pid_struct(pid);
}

u32 __attribute__((always_inline)) get_namespace_nr_from_task_struct(struct task_struct *task) {
    struct pid *pid = NULL;
    bpf_probe_read_kernel(&pid, sizeof(pid), (void *)task + get_task_struct_pid_offset());

    u32 pid_level = 0;
    bpf_probe_read_kernel(&pid_level, sizeof(pid_level), (void *)pid + get_pid_level_offset());

    // read the namespace nr from &pid->numbers[pid_level].nr
    u32 namespace_nr = 0;
    u64 namespace_numbers_offset = pid_level * get_sizeof_upid();
    bpf_probe_read_kernel(&namespace_nr, sizeof(namespace_nr), (void *)pid + get_pid_numbers_offset() + namespace_numbers_offset);

    return namespace_nr;
}

__attribute__((always_inline)) struct process_event_t *new_process_event(u8 is_fork) {
    u32 key = bpf_get_current_pid_tgid() % EVENT_GEN_SIZE;
    struct process_event_t *evt = bpf_map_lookup_elem(&process_event_gen, &key);

    if (evt) {
        __builtin_memset(evt, 0, sizeof(*evt));
        if (!is_fork) {
            evt->event.flags |= EVENT_FLAGS_ACTIVITY_DUMP_SAMPLE;
        }
    }

    return evt;
}

bool __attribute__((always_inline)) is_current_kworker_dying() {
    char comm[16];
    bpf_get_current_comm(comm, sizeof(comm));
    return comm[0] == 'k' && comm[1] == 'w' && comm[2] == 'o' && comm[3] == 'r' && comm[4] == 'k' && comm[5] == 'e' && comm[6] == 'r' && comm[7] == '/' && comm[8] == 'd' && comm[9] == 'y' && comm[10] == 'i' && comm[11] == 'n' && comm[12] == 'g';
}

#endif
