// Vian Chen <imvianchen@stu.pku.edu.cn>
// User event module

#include <linux/eventfd.h>
#include "micom.h"

struct eventfd_ctx *ctxp = NULL;

void uevent_send(int code)
{
    int command_ind = -1;

    if (!ctxp)
        return;

    for (int i = 0; i < MAX_CODE; i++) {
        if (code == desc_ind[i]) {
            command_ind = i;
            micom_info("sending: %s", comm_desc[i]);
            break;
        }
    }

    if (command_ind < 0) {
        micom_err("invalid code: %d", code);
        return;
    }

    // As this is a very simplified mock environment, it's very
    // unlikely that the counter would overflow.
    eventfd_signal(ctxp, code);
}

void uevent_unregister(void)
{
    if (ctxp) {
        eventfd_ctx_put(ctxp);
    }
}

int uevent_register(int ueventfd)
{
    micom_info("ueventfd from uspace: %d", ueventfd);

    // By design, only the most recently opened eventfd
    // client will get messages
    ctxp = eventfd_ctx_fdget(ueventfd);
    if (IS_ERR(ctxp)) {
        micom_err("failed to get eventfd context");
        return -EINVAL;
    }

    return 0;
}
