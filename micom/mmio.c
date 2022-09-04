// Vian Chen <imvianchen@stu.pku.edu.cn>

#include <linux/mm.h>
#include <linux/kthread.h>
#include <linux/delay.h>
#include <linux/sched.h>
#include <asm-generic/io.h>
#include "micom.h"

#define MMIO_BUFF_SIZE 0x1000
#define POLL_INTERVAL 100000

static struct task_struct *mmiot;
static uint32_t *buff;

int mmio_mmap(struct file *file, struct vm_area_struct *vma)
{
    unsigned long size = vma->vm_end - vma->vm_start;
    unsigned long offset = vma->vm_pgoff << PAGE_SHIFT;
    unsigned long pfn = vma->vm_pgoff + (virt_to_phys(buff) >> PAGE_SHIFT);

    micom_info("size = %lu, offset = %lu, pfn = %lu", size, offset, pfn);

    if (offset + size > MMIO_BUFF_SIZE)
        return -EINVAL;

    vma->vm_page_prot = pgprot_noncached(vma->vm_page_prot);
    return remap_pfn_range(vma, vma->vm_start, pfn + (offset >> PAGE_SHIFT),
        size, vma->vm_page_prot);
}

static int listen_for_mmio_input(void * /* unused */ args)
{
    while (!kthread_should_stop()) {
        if (buff[0] != 0) {
            micom_info("mmiot detected change. code: %d", buff[0]);
            uevent_send(buff[0]);
            buff[0] = 0;
        }
        usleep_range(POLL_INTERVAL, POLL_INTERVAL);
    }

    return 0;
}

int mmio_register(void)
{
    buff = kvmalloc(MMIO_BUFF_SIZE, GFP_KERNEL);
    if (!buff) {
        pr_err("failed allocating page");
        goto err_out;
    }

    mmiot = kthread_create(listen_for_mmio_input, NULL, "mmiot");
    if (!mmiot) {
        pr_err("creating mmiot failed");
        goto err_out;
    }

    wake_up_process(mmiot);
    return 0;

err_out:
    if (buff)
        kvfree(buff);
    
    return -ENOMEM;
}

void mmio_unregister(void)
{
    if (buff)
        kvfree(buff);
    
    if (mmiot) {
        micom_info("mmio listening is about to stop");
        kthread_stop(mmiot);
    }
}
