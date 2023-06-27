import os
import re
import subprocess as sp
import sys
import textwrap
import threading
import time
import unicodedata

from enum import Enum

BranchStatus = Enum('BranchState', ['even', 'ahead', 'behind', 'diverged'])


class GitError(Exception):
    def __init__(self, message, stderr=None):
        if stderr:
            stderr = stderr.strip()
            if '\n' in stderr:
                message += ':\n' + textwrap.indent(stderr, '  ')
            else:
                message += ': "{}"'.format(stderr)
        super().__init__(message)


class GitBranchDiverged(GitError):
    def __init__(self, branch1, branch2):
        super().__init__('{} has diverged from {}'.format(branch1, branch2))


class GitRevUnrelated(GitError):
    def __init__(self, branch1, branch2):
        super().__init__('{} is unrelated to {}'.format(branch1, branch2))


class GitCommitError(GitError):
    def __init__(self, stderr):
        super().__init__('error writing commit', stderr)


class GitEmptyCommit(GitError):
    def __init__(self):
        super().__init__('nothing changed in commit')


def git_path(path):
    path = os.path.normpath(path)
    # Fix unicode pathnames on macOS; see
    # <http://stackoverflow.com/a/5582439/44289>.
    if sys.platform == 'darwin':  # pragma: no cover
        if isinstance(path, bytes):
            path = path.decode('utf-8')
        path = unicodedata.normalize('NFKC', path)
    return '/'.join(path.split(os.path.sep))


def make_when(timestamp=None):
    if timestamp is None:
        timestamp = int(time.time())
    timezone = '{:+05d}'.format(-1 * time.timezone // 3600 * 100)
    return '{} {}'.format(timestamp, timezone)


def get_config(key, encoding='utf-8'):
    cmd = ['git', 'config', key]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, encoding=encoding)
    if p.returncode != 0:
        raise GitError('error getting config {!r}'.format(key), p.stderr)
    return p.stdout.strip()


def get_commit_encoding():
    try:
        return get_config('i18n.commitEncoding')
    except GitError:
        return 'utf-8'


def get_latest_commit(rev, *, short=False):
    cmd = ['git', 'rev-parse'] + (['--short'] if short else []) + [rev]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        raise GitError('error getting latest commit', p.stderr)
    return p.stdout.strip()


def count_reachable(rev):
    cmd = ['git', 'rev-list', '--count', rev]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode == 0:
        return int(p.stdout.strip())
    raise GitError('unable to get number of reachable commits from {}'
                   .format(rev), p.stderr)


def get_ref(branch, *, nonexist_ok=False):
    cmd = ['git', 'rev-parse', '--symbolic-full-name', branch]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        if nonexist_ok:
            return 'refs/heads/{}'.format(branch)
        raise GitError('error getting git ref for {}'.format(branch), p.stderr)
    return p.stdout.strip()


def update_ref(branch, new_ref):
    cmd = ['git', 'update-ref', get_ref(branch, nonexist_ok=True), new_ref]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        raise GitError('error updating ref for {}'.format(branch), p.stderr)


def has_branch(branch):
    try:
        get_latest_commit(branch)
        return True
    except GitError:
        return False


def get_merge_base(rev1, rev2):
    cmd = ['git', 'merge-base', rev1, rev2]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)

    if p.returncode == 0:
        return p.stdout.strip()
    elif p.returncode == 1:
        raise GitRevUnrelated(rev1, rev2)
    raise GitError('error getting merge-base', p.stderr)


def compare_branches(branch1, branch2):
    base = get_merge_base(branch1, branch2)
    latest1 = get_latest_commit(branch1)
    latest2 = get_latest_commit(branch2)

    if base == latest1:
        return BranchStatus.even if base == latest2 else BranchStatus.behind
    else:
        return BranchStatus.ahead if base == latest2 else BranchStatus.diverged


def update_from_upstream(remote, branch):
    remote_branch = '{}/{}'.format(remote, branch)
    if not has_branch(remote_branch):
        return

    if not has_branch(branch):
        update_ref(branch, get_latest_commit(remote_branch))
    else:
        status = compare_branches(branch, remote_branch)
        if status == BranchStatus.behind:
            update_ref(branch, get_latest_commit(remote_branch))
        if status == BranchStatus.diverged:
            raise GitBranchDiverged(branch, remote_branch)


def push_branch(remote, branch):
    cmd = ['git', 'push', '--', remote, branch]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        raise GitError('failed to push branch {} to {}'.format(branch, remote),
                       p.stderr)


def delete_branch(branch):
    cmd = ['git', 'branch', '--delete', '--force', branch]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        raise GitError('unable to delete branch {}'.format(branch),
                       p.stderr)


def is_commit_empty(rev):
    cmd = ['git', 'log', '-1', '--format=', '--name-only', rev]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode == 0:
        return not p.stdout
    raise GitError('error getting commit changes', p.stderr)


def delete_latest_commit(branch):
    if count_reachable(branch) > 1:
        update_ref(branch, get_latest_commit(branch + '^'))
    else:
        delete_branch(branch)


class FileInfo:
    def __init__(self, path, data, mode=0o100644):
        if isinstance(data, str):
            data = data.encode('utf-8')
        self.path = path
        self.data = data
        self.mode = mode

    def __eq__(self, rhs):
        return (self.path == rhs.path and self.data == rhs.data and
                self.mode == rhs.mode)

    def __repr__(self):
        return '<FileInfo({!r}, {:06o})>'.format(self.path, self.mode)

    def copy(self, destdir='', start=''):
        return FileInfo(
            os.path.join(destdir, os.path.relpath(self.path, start)),
            self.data, self.mode
        )


class Commit:
    def __init__(self, branch, message, *, allow_empty=False):
        cmd = ['git', 'fast-import', '--date-format=rfc2822', '--quiet',
               '--done']
        self._pipe = sp.Popen(cmd, stdin=sp.PIPE, stderr=sp.PIPE,
                              universal_newlines=False)
        self._finished = False
        self._allow_empty = allow_empty

        self._stderr = b''
        self._read_thread = threading.Thread(target=self._read)
        self._read_thread.start()

        try:
            self._start_commit(branch, message)
        except Exception:
            self.abort()
            raise

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        if not self._finished:
            if exc_type:
                self.abort()
            else:
                self.finish()

    @staticmethod
    def _escape_path(path):
        if re.search(r'[\n\"]', path):
            return '"' + re.sub(r'[\n\"\\]', r'\\\g<0>', path) + '"'
        return path

    def _read(self):
        while True:
            line = self._pipe.stderr.readline()
            if not line:
                break
            self._stderr += line
        self._pipe.stderr.close()

    def _write(self, data):
        if isinstance(data, str):
            data = data.encode('utf-8')
        try:
            return self._pipe.stdin.write(data)
        except BrokenPipeError:  # pragma: no cover
            raise GitCommitError(self._stderr.decode('utf-8'))

    def _write_data(self, data):
        if isinstance(data, str):
            data = data.encode('utf-8')
        self._write('data {}\n'.format(len(data)))
        self._write(data)
        self._write('\n')

    def _start_commit(self, branch, message):
        self._branch = branch
        encoding = get_commit_encoding()

        name = (os.getenv('GIT_COMMITTER_NAME') or
                get_config('user.name', encoding))
        name = re.sub(r'[<>\n]', '', name)

        email = (os.getenv('GIT_COMMITTER_EMAIL') or
                 get_config('user.email', encoding))
        email = re.sub(r'[<>\n]', '', email)

        when = os.getenv('GIT_COMMITTER_DATE') or make_when()

        self._write('commit {}\n'.format(get_ref(branch, nonexist_ok=True)))
        self._write('committer {name}<{email}> {time}\n'.format(
            name=name + ' ' if name else '', email=email, time=when
        ))
        self._write_data(message)
        try:
            head = get_latest_commit(branch)
            self._write('from {}\n'.format(head))
        except GitError:
            pass

    def delete_files(self, files):
        if files == '*':
            self._write('deleteall\n')
        else:
            for f in files:
                self._write('D {}\n'.format(self._escape_path(git_path(f))))

    def add_file(self, file_info):
        self._write('M {mode:06o} inline {path}\n'.format(
            path=self._escape_path(git_path(file_info.path)),
            mode=file_info.mode
        ))
        self._write_data(file_info.data)

    def finish(self):
        if self._finished:
            raise GitError('commit already finalized')
        self._finished = True

        self._write('done\n')
        self._pipe.stdin.close()
        self._read_thread.join()
        if self._pipe.wait() != 0:
            raise GitCommitError(self._stderr.decode('utf-8'))

        if ( not self._allow_empty
             and is_commit_empty(get_latest_commit(self._branch)) ):
            delete_latest_commit(self._branch)
            raise GitEmptyCommit()

    def abort(self):
        if self._finished:
            raise GitError('commit already finalized')
        self._finished = True

        try:
            self._pipe.stdin.close()
        except BrokenPipeError:  # pragma: no cover
            pass
        self._pipe.terminate()
        self._read_thread.join()
        self._pipe.wait()


def real_path(branch, filename):
    path = ''
    for i in git_path(filename).split('/'):
        if path:
            path += '/'
        curr_path = path + i
        mode = file_mode(branch, curr_path, follow_symlinks=False)
        if mode == 0o120000:
            curr_path = path + read_file(branch, curr_path,
                                         universal_newlines=True,
                                         follow_symlinks=False)
        path = curr_path
    return path


def file_mode(branch, filename, follow_symlinks=True):
    filename = filename.rstrip('/')
    # The root directory of the repo is, well... a directory.
    if not filename:
        return 0o040000

    if follow_symlinks:
        filename = real_path(branch, filename)

    cmd = ['git', 'ls-tree', '--full-tree', '--', branch, git_path(filename)]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE, universal_newlines=True)
    if p.returncode != 0:
        raise GitError('unable to read file {!r}'.format(filename), p.stderr)
    if not p.stdout:
        raise GitError('file not found')

    return int(p.stdout.split(' ', 1)[0], 8)


def read_file(branch, filename, universal_newlines=False,
              follow_symlinks=True):
    if follow_symlinks:
        filename = real_path(branch, filename)

    cmd = ['git', 'show', '{branch}:{filename}'.format(
        branch=branch, filename=git_path(filename)
    )]
    p = sp.run(cmd, stdout=sp.PIPE, stderr=sp.PIPE,
               universal_newlines=universal_newlines)
    if p.returncode != 0:
        raise GitError('unable to read file {!r}'.format(filename),
                       str(p.stderr))
    return p.stdout


def walk_files(branch, path=''):
    gpath = git_path(path) if path else ''
    cmd = ['git', 'ls-tree', '--full-tree', '-r', '--',
           '{branch}:{path}'.format(branch=branch, path=gpath)]
    p = sp.Popen(cmd, stdout=sp.PIPE, stderr=sp.DEVNULL,
                 universal_newlines=True)

    for line in p.stdout:
        strmode, _, _, filename = re.split(r'\s', line.rstrip(), 3)
        mode = int(strmode, 8)
        filepath = os.path.join(path, os.path.normpath(filename))
        yield FileInfo(filepath, read_file(branch, filepath), mode)

    p.stdout.close()
    if p.wait() != 0:
        # It'd be nice if we could read from stderr, but it's somewhat
        # complex to do that while avoiding deadlocks. (select(2) does this
        # on POSIX systems, but that doesn't work on Windows.)
        raise GitError("unable to read files in '{branch}:{path}'"
                       .format(branch=branch, path=gpath))


def walk_real_files(srcdir):
    for path, dirs, filenames in os.walk(srcdir):
        if '.git' in dirs:
            dirs.remove('.git')
        for f in filenames:
            filepath = os.path.join(path, f)
            mode = 0o100755 if os.access(filepath, os.X_OK) else 0o100644
            with open(filepath, 'rb') as fd:
                data = fd.read()
            yield FileInfo(filepath, data, mode)
