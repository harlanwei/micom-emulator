#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>
#include <sys/eventfd.h>
#include <sys/ioctl.h>
#include "include/refcodes.h"

static int efd;
static int retval;
static fd_set rfds;

#define err(format, ...) \
    fprintf(stderr, "watchdog: " format "\n", ##__VA_ARGS__)

int main(int argc, char **argv) {
    int s, micomfd;
    uint64_t ctr = 0;

    efd = eventfd(0, 0);
    if (efd < 0) {
        err("cannot create eventfd");
        return -1;
    }

    FD_ZERO(&rfds);
    FD_SET(efd, &rfds);

    micomfd = open("/dev/micom", O_WRONLY);
    if (micomfd < 0) {
        err("cannot open /dev/micom");
        return -1;
    }
    ioctl(micomfd, _IOC(IOC_IN, 0x15, 0, sizeof(efd)), efd);
    close(micomfd);

    while (1) {
        retval = select(efd + 1, &rfds, NULL, NULL, NULL);
        if (retval < 0) {
            err("select failed");
            return -1;
        } else if (retval > 0) {
            s = read(efd, &ctr, sizeof(ctr));
            if (s < sizeof(ctr)) {
                err("eventfd read error");
                return -1;
            }
            printf("COMMAND EXECUTED: %s\n", comm_desc[ctr]);
        }
    }

    close(efd);
    return 0;
}
