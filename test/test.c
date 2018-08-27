#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "libipfs.h"

int added = -1;

void file_added(char* error, char* data, size_t size) {
    if (data != NULL) {
        fprintf(stdout, "hash: %s\n", data);
        free(data);
    }

    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        free(error);
        added = 1;
    }

    ipfs_stop();
    added = 0;
}

int main(int argc, char **argv) {
    char *path, *err;

    if (argc < 2) {
        fprintf(stderr, "missing argument\n");
        return 1;
    }

    path = strdup(argv[1]);
    err = ipfs_start(path);
    free(path);

    if (err != NULL) {
        fprintf(stderr, "error: %s\n", err);
        free(err);
        return 1;
    }

    char data[] = "content for ipfs wrapper test";

    ipfs_add((void*)data, sizeof(data), (void*)&file_added);

    while(added == -1) {};

    return added;
}

