#include <linux/proc_fs.h>
#include "micom.h"

#define MAX_SIGNAL_BUFFER 3

static struct proc_dir_entry *micom_dir, *signal_entry;

static ssize_t signal_write(struct file *filp, const char __user *buf, size_t count, loff_t *f_pos)
{
    int code;
    char *endp;

    if (count > MAX_SIGNAL_BUFFER) {
        micom_err("invalid code");
        return -EINVAL;
    }

    code = simple_strtol(buf, &endp, 10);
    if (endp == buf) {
        micom_err("invalid code");
        return -EINVAL;
    }

    if (code < 0) {
        micom_err("invalid code: %d", code);
        return -EINVAL;
    }

    uevent_send(code);

    return count;
}

static struct proc_ops micom_fops = {
    .proc_write = signal_write,
};

int proc_register(void)
{
    micom_dir = proc_mkdir("micom", NULL);
    if (micom_dir == NULL) {
        micom_err("failed to create proc dir");
        return -EIO;
    }

    signal_entry = proc_create("signal", 0666, micom_dir, &micom_fops);
    if (signal_entry == NULL) {
        micom_err("failed to create signal entry");
        remove_proc_entry("micom", NULL);
        return -EIO;
    }

    return 0;
}

void proc_unregister(void)
{
    if (signal_entry) {
        remove_proc_entry("signal", micom_dir);
    }

    if (micom_dir) {
        remove_proc_entry("micom", NULL);
    }

    micom_info("procfs unregistered");
}
