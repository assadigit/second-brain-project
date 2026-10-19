// webkit-path-shim.c — the rootless-webkit runtime shim (T-70's Linux
// evidence leg; the same 30-line trick SESSION-001 used, committed this
// time so the recipe is reproducible from the repo).
//
// WebKit 2.54 bakes its libexec prefix (/usr/lib/x86_64-linux-gnu/
// webkit2gtk-4.1) into the library: it spawns WebKitNetworkProcess/
// WebKitWebProcess/WebKitGPUProcess and dlopens the injected bundle from
// that absolute path — and WEBKIT_EXEC_PATH alone is not honored. This box
// has no system webkit (the runtime lives only in the rootless tree at
// ~/.local/lib-wails), so the shim rewrites every access to the compile-
// time prefix onto WEBKIT_TREE_LIBEXEC. LD_PRELOADed into the GUI process;
// child processes inherit it through the environment.
//
// Build: gcc -shared -fPIC -o webkit-path-shim.so webkit-path-shim.c -ldl
#define _GNU_SOURCE
#include <dlfcn.h>
#include <spawn.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static const char SYS_PREFIX[] = "/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1";
static const char *tree_libexec;

__attribute__((constructor)) static void shim_init(void) {
    tree_libexec = getenv("WEBKIT_TREE_LIBEXEC");
}

// rewrite returns `path` unchanged unless it sits under the compile-time
// prefix — then a tree-anchored twin in `buf` (or `path` on overflow).
static const char *rewrite(const char *path, char *buf, size_t bufsz) {
    if (!path || !tree_libexec) return path;
    const size_t n = sizeof SYS_PREFIX - 1;
    if (strncmp(path, SYS_PREFIX, n) != 0) return path;
    if (path[n] != '/' && path[n] != '\0') return path;
    const int w = snprintf(buf, bufsz, "%s%s", tree_libexec, path + n);
    if (w < 0 || (size_t)w >= bufsz) return path;
    return buf;
}

int execve(const char *path, char *const argv[], char *const envp[]) {
    static int (*real)(const char *, char *const[], char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "execve");
    char buf[4096];
    return real(rewrite(path, buf, sizeof buf), argv, envp);
}

int open(const char *path, int flags, ...) {
    static int (*real)(const char *, int, ...);
    if (!real) real = dlsym(RTLD_NEXT, "open");
    va_list ap;
    va_start(ap, flags);
    mode_t mode = va_arg(ap, mode_t);
    va_end(ap);
    char buf[4096];
    return real(rewrite(path, buf, sizeof buf), flags, mode);
}

int openat(int dirfd, const char *path, int flags, ...) {
    static int (*real)(int, const char *, int, ...);
    if (!real) real = dlsym(RTLD_NEXT, "openat");
    va_list ap;
    va_start(ap, flags);
    mode_t mode = va_arg(ap, mode_t);
    va_end(ap);
    char buf[4096];
    return real(dirfd, rewrite(path, buf, sizeof buf), flags, mode);
}

int stat(const char *path, struct stat *st) {
    static int (*real)(const char *, struct stat *);
    if (!real) real = dlsym(RTLD_NEXT, "stat");
    char buf[4096];
    return real(rewrite(path, buf, sizeof buf), st);
}

int access(const char *path, int mode) {
    static int (*real)(const char *, int);
    if (!real) real = dlsym(RTLD_NEXT, "access");
    char buf[4096];
    return real(rewrite(path, buf, sizeof buf), mode);
}

// glib's g_spawn_async rides posix_spawn (not execve) — the daemons'
// spawn path on this box.
int posix_spawn(pid_t *pid, const char *path,
                const posix_spawn_file_actions_t *file_actions,
                const posix_spawnattr_t *attrp,
                char *const argv[], char *const envp[]) {
    static int (*real)(pid_t *, const char *,
        const posix_spawn_file_actions_t *, const posix_spawnattr_t *,
        char *const[], char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "posix_spawn");
    char buf[4096];
    return real(pid, rewrite(path, buf, sizeof buf), file_actions, attrp, argv, envp);
}

int posix_spawnp(pid_t *pid, const char *file,
                 const posix_spawn_file_actions_t *file_actions,
                 const posix_spawnattr_t *attrp,
                 char *const argv[], char *const envp[]) {
    static int (*real)(pid_t *, const char *,
        const posix_spawn_file_actions_t *, const posix_spawnattr_t *,
        char *const[], char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "posix_spawnp");
    char buf[4096];
    return real(pid, rewrite(file, buf, sizeof buf), file_actions, attrp, argv, envp);
}

void *dlopen(const char *filename, int flags) {
    static void *(*real)(const char *, int);
    if (!real) real = dlsym(RTLD_NEXT, "dlopen");
    char buf[4096];
    return real(rewrite(filename, buf, sizeof buf), flags);
}
