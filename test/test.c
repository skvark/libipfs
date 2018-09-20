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
int _stats_done = -1;
int _config_done = -1;
int _files_ls_done = -1;
int _files_cp_done = -1;
int _files_mkdir_done = -1;
int _files_stat_done = -1;
int started = -1;

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

void config_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "config: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        _config_done = 1;
    }
    _config_done = 0;
}

void start_done(char* error, char* data, size_t size, int f, void* instance) {
    if (error != NULL) {
        fprintf(stderr, "error: %s\n", error);
        started = 1;
    }
    started = 0;
}

void stats_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "stats: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "stats error: %s\n", error);
        _stats_done = 1;
    }
    _stats_done = 0;
}

void files_cp_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "files cp: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "files cp error: %s\n", error);
        _files_cp_done = 1;
    }
    _files_cp_done = 0;
}

void files_ls_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "files ls: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "files ls error: %s\n", error);
        _files_ls_done = 1;
    }
    _files_ls_done = 0;
}

void files_mkdir_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "mkdir: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "mkdir error: %s\n", error);
        _files_mkdir_done = 1;
    }
    _files_mkdir_done = 0;
}

void files_stat_done(char* error, char* data, size_t size, int f, void* instance) {
    if (data != NULL) {
        fprintf(stdout, "files stat: %s\n", data);
    }

    if (error != NULL) {
        fprintf(stderr, "files stat error: %s\n", error);
        _files_stat_done = 1;
    }
    _files_stat_done = 0;
}

int main(int argc, char **argv) {
    char *path;

    if (argc < 2) {
        fprintf(stderr, "missing argument\n");
        return 1;
    }

    path = strdup(argv[1]);
    char repo_max_size[] = "500MB";

    ipfs_start(path, (char*)repo_max_size, (void*)&start_done);

    while(started == -1) {}

    char data[] = "content for ipfs wrapper test";
    char pathname[] = "test_folder";
    char mfsdir[] = "/mfs_test";
    char mfsfile[] = "/mfs_test/test.txt";
    char cat_file[] = "QmVno5qCuKwMt5wB1X6NUm4obzqhHbvBhkswkV3VXC3k4s";
    char mfs_cp[] = "/ipfs/QmXvrr3gPtddcNrisH7i2nan9rY7v7RcxVQ9jjRreoWwRS";
    char ls_path[] = "QmU1BMQS1hPgLziefeC2tWm3STK3hL8nMrDCX27VcwKuDm";

    ipfs_add_bytes((void*)data, sizeof(data), (void*)&file_added);
    ipfs_add_path_or_file((char*)pathname, (void*)&file_path_added);

    ipfs_id(NULL, (void*)&id_done);
    ipfs_peers((void*)&peers_done);
    ipfs_config((void*)&config_done);
    ipfs_repo_stats((void*)&stats_done);

    while(added == -1 || path_added == -1 || _peers_done == -1 || _id_done == -1 || _config_done == -1 || _stats_done == -1) {};

    ipfs_cat((char*)cat_file, (void*)cat_done);
    ipfs_ls((char*)ls_path, (void*)ls_done);

    while(_cat_done == -1 || _ls_done == -1) {};

    ipfs_files_mkdir((char*)mfsdir, 1, (void*)files_mkdir_done);

    while(_files_mkdir_done == -1) {};

    ipfs_files_cp((char*)mfs_cp, (char*)mfsfile, (void*)files_cp_done);

    while(_files_cp_done == -1) {};

    ipfs_files_ls(NULL, (void*)files_ls_done);
    ipfs_files_stat((char*)mfsfile, (void*)files_stat_done);

    while(_files_ls_done == -1 || _files_stat_done == -1) {};

    _files_ls_done = -1;

    ipfs_files_ls((char*)mfsdir, (void*)files_ls_done);

    while(_files_ls_done == -1) {};

    ipfs_stop();

    return added && path_added;
}

