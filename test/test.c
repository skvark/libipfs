#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "libipfs.h"

int added = -1;
int path_added = -1;
int _unpin_done = -1;
int _cat_done = -1;
int _ls_done = -1;
int _peers_done = -1;
int _id_done = -1;

void file_added(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "hash: %s\n", data);
    }
    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        added = 1;
    }
    added = 0;
}

void ls_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "ls: %s\n", data);
    }
    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _ls_done = 1;
    }
    _ls_done = 0;
}

void file_path_added(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "hash (msf root): %s\n", data);
    }
    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        path_added = 1;
    }
    path_added = 0;
}

void cat_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "cat data: %s\n", data);
    }
    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _cat_done = 1;
    }
    _cat_done = 0;
}

void unpin_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "unpinned: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _unpin_done = 1;
    }
    _unpin_done = 0;
}

void peers_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "peers: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _peers_done = 1;
    }
    _peers_done = 0;
}

void id_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "id: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _id_done = 1;
    }
    _id_done = 0;
}

int main(int argc, char **argv) {
    char *path, *err;

    if (argc < 2) {
        fprintf(stderr, "missing argument\n");
        return 1;
    }

    path = strdup(argv[1]);
    err = ipfs_start(path);

    if (err != NULL) {
        fprintf(stderr, "error: %s\n", err);
        return 1;
    }

    char data[] = "content for ipfs wrapper test";
    char pathname[] = "test_folder";
    char cat_file[] = "QmVno5qCuKwMt5wB1X6NUm4obzqhHbvBhkswkV3VXC3k4s";
    char ls_path[] = "QmU1BMQS1hPgLziefeC2tWm3STK3hL8nMrDCX27VcwKuDm";

    ipfs_add_bytes((void*)data, sizeof(data), (void*)&file_added);
    ipfs_add_path_or_file((char*)pathname, (void*)&file_path_added);

    ipfs_id(NULL, (void*)&id_done);
    ipfs_peers((void*)&peers_done);

    while(added == -1 || path_added == -1 || _peers_done == -1 || _id_done == -1) {};

    ipfs_cat((char*)cat_file, (void*)cat_done);
    ipfs_ls((char*)ls_path, (void*)ls_done);

    while(_cat_done == -1 || _ls_done == -1) {};

    ipfs_stop();

    return added && path_added;
}

