// Vian Chen <imvianchen@stu.pku.edu.cn>

#ifndef MODULE_MICOM_H
#define MODULE_MICOM_H

#include <linux/module.h>
#include <linux/init.h>
#include "refcodes.h"

#define DEVICE_NAME "micom"

#define MODULE_PREFIX DEVICE_NAME
#define MICOM_KERNEL_MESSAGE(kind, format, ...) \
    pr_##kind( \
        MODULE_PREFIX ": " format " (in %s at line %d)\n", \
        ##__VA_ARGS__, __func__, __LINE__)
#define micom_info(format, ...) MICOM_KERNEL_MESSAGE(info, format, ##__VA_ARGS__)
#define micom_err(format, ...) MICOM_KERNEL_MESSAGE(err, format, ##__VA_ARGS__)
#define micom_warn(format, ...) MICOM_KERNEL_MESSAGE(warn, format, ##__VA_ARGS__)

/// uevent.c
extern struct eventfd_ctx *ctxp;

void uevent_send(int code);
void uevent_unregister(void);
int uevent_register(int ueventfd);

/// proc.c
int proc_register(void);
void proc_unregister(void);

/// mmio.c
extern struct mmio mmio;
int mmio_mmap(struct file *file, struct vm_area_struct *vma);
int mmio_register(void);
void mmio_unregister(void);

#endif // MODULE_MICOM_H
