// Vian Chen <imvianchen@stu.pku.edu.cn>

#include <linux/fs.h>
#include <linux/device.h>
#include <linux/eventfd.h>
#include "micom.h"

static int major;
static dev_t devno;
static struct class *device_class;

static int micom_open(struct inode *inode, struct file *filp)
{
    return 0;
}

/**
 * To simplify the implemetation, the python script is responsible for encoding
 * the original commands into ref codes.
 */
static long micom_ioctl(struct file *filp, unsigned int cmd, unsigned long param)
{
    int type, number;
    int ret = 0;

    type = _IOC_TYPE(cmd);
    if (type != 0x15) {
        /* not directed to this driver. do nothing. */
        return 0;
    }
    
    number = _IOC_NR(cmd);
    micom_info("ioctl: number = %d, param = %lu", number, param);

    if (number == 0) {
        ret = uevent_register((int) param);
    } else {
        uevent_send(number);
    }

    return ret;
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
    .mmap = mmio_mmap,
};

static int __init init_micom(void)
{
    struct device *p_device;

    micom_info("initializing micom driver");

    major = register_chrdev(0, DEVICE_NAME, &fops);
    if (major < 0) {
        micom_err("failed registering device. ret = %d", major);
        return major;
    }

    devno = MKDEV(major, 0);
    device_class = class_create(THIS_MODULE, DEVICE_NAME);
    if (IS_ERR(device_class)) {
        micom_err("can't create class");
        goto destroy_chrdev;
    }
    device_class->devnode = micom_devnode;

    p_device = device_create(device_class, NULL, devno, NULL, DEVICE_NAME);
    if (IS_ERR((p_device))) {
        micom_err("can't create device file");
        goto destroy_class;
    }

    if (proc_register()) {
        micom_err("failed creating procfs entries");
        goto destroy_device;
    }

    if (mmio_register()) {
        micom_err("failed registering mmio");
        goto destroy_proc;
    }

    micom_info("successfully loaded");
    return 0;

destroy_proc:
    proc_unregister();
destroy_device:
    device_destroy(device_class, devno);
destroy_class:
    class_destroy(device_class);
destroy_chrdev:
    unregister_chrdev_region(devno, 1);
    return -1;
}

static void __exit exit_micom(void)
{
    uevent_unregister();
    proc_unregister();
    mmio_unregister();

    device_destroy(device_class, devno);
    class_destroy(device_class);
    unregister_chrdev(major, DEVICE_NAME);
}

module_init(init_micom);
module_exit(exit_micom);

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Vian Chen <imvianchen@stu.pku.edu.cn>");
