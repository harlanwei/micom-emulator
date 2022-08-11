// Vian Chen <imvianchen@stu.pku.edu.cn>

#include <linux/module.h>
#include <linux/fs.h>
#include <linux/device.h>
#include <linux/eventfd.h>
#include "../include/refcodes.h"

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Vian Chen <imvianchen@stu.pku.edu.cn>");

#define DEVICE_NAME "micom"

#define MODULE_PREFIX DEVICE_NAME
#define MICOM_KERNEL_MESSAGE(kind, format, ...) \
    pr_##kind( \
        MODULE_PREFIX ": " format " (in %s at line %d)\n", \
        ##__VA_ARGS__, __func__, __LINE__)
#define micom_info(format, ...) MICOM_KERNEL_MESSAGE(info, format, ##__VA_ARGS__)
#define micom_err(format, ...) MICOM_KERNEL_MESSAGE(err, format, ##__VA_ARGS__)
#define micom_warn(format, ...) MICOM_KERNEL_MESSAGE(warn, format, ##__VA_ARGS__)

static int major;
static dev_t devno;
static struct class *device_class;
static struct eventfd_ctx *ctxp;


static int micom_open(struct inode *inode, struct file *filp)
{
    return 0;
}

static void comm_exec(int code)
{
    if (!ctxp)
        return;

    // As this is a very simplified mock environment, it's very
    // unlikely that the counter would overflow.
    eventfd_signal(ctxp, code);
}

/**
 * To simplify the implemetation, the python script is responsible for encoding
 * the original commands into ref codes.
 */
static long micom_ioctl(struct file *filp, unsigned int cmd, unsigned long param)
{
    int type, number;
    micom_info("ioctl: cmd = %d, param = %lu", cmd, param);

    type = _IOC_TYPE(cmd);
    if (type != 0x15) {
        /* not directed to this driver. do nothing. */
        return 0;
    }
    
    number = _IOC_NR(cmd);
    if (number < 0 || number > MAX_CODE) {
        micom_err("invalid code: %d", number);
    }

    if (number == 0) {
        int ueventfd = (int) param;
        micom_info("ueventfd from uspace: %d", ueventfd);

        ctxp = eventfd_ctx_fdget(ueventfd);
        if (IS_ERR(ctxp)) {
            micom_err("failed to get eventfd context");
            return -EINVAL;
        }
    } else {
        micom_info("sending: %s", comm_desc[number]);
        comm_exec(number);
    }

    return 0;
}

static int micom_release(struct inode *inode, struct file *filp)
{
    return 0;
}

static char *micom_devnode(struct device *dev, umode_t *mode)
{
    if (!mode)
        return NULL;
    if (dev->devt == devno)
        *mode = 0666;
    return NULL;
}

static struct file_operations fops = {
    .open = micom_open,
    .unlocked_ioctl = micom_ioctl,
    .release = micom_release,
};

static int __init init_micom(void)
{
    struct device *p_device;

    major = register_chrdev(0, DEVICE_NAME, &fops);
    if (major < 0) {
        micom_err("failed registering device. ret = %d", major);
        return major;
    }

    devno = MKDEV(major, 0);
    device_class = class_create(THIS_MODULE, DEVICE_NAME);
    if (IS_ERR(device_class)) {
        micom_err("can't create class");
        goto err_class_create;
    }
    device_class->devnode = micom_devnode;

    p_device = device_create(device_class, NULL, devno, NULL, DEVICE_NAME);
    if (IS_ERR((p_device))) {
        micom_err("can't create device file");
        goto err_device_create;
    }

    micom_info("successfully loaded");
    return 0;

err_device_create:
    class_destroy(device_class);

err_class_create:
    unregister_chrdev_region(devno, 1);
    return -1;
}

static void __exit exit_micom(void)
{
    device_destroy(device_class, devno);
    class_destroy(device_class);
    unregister_chrdev(major, DEVICE_NAME);

    if (ctxp) {
        eventfd_ctx_put(ctxp);
    }
}

module_init(init_micom);
module_exit(exit_micom);
