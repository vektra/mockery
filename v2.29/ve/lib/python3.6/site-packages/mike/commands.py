import importlib_resources as resources
import http.server
import os
import posixpath
from contextlib import contextmanager
from enum import Enum
from jinja2 import Template

from . import git_utils
from . import mkdocs_utils
from . import server
from .app_version import version as app_version
from .versions import Versions

versions_file = 'versions.json'
AliasType = Enum('AliasType', ['symlink', 'copy', 'redirect'])


def _format_deploy_prefix(deploy_prefix):
    return ' in {}'.format(deploy_prefix) if deploy_prefix else ''


def _redirect_template(user_template=None):
    template_file = (
        user_template or
        resources.files('mike').joinpath('templates/redirect.html')
    )
    with open(template_file, 'rb') as f:
        return Template(f.read().decode('utf-8'), autoescape=True,
                        keep_trailing_newline=True)


def _add_redirect_to_commit(commit, template, src, dst,
                            use_directory_urls):
    if os.path.splitext(src)[1] == '.html':
        reldst = os.path.relpath(dst, os.path.dirname(src))
        href = '/'.join(reldst.split(os.path.sep))
        if use_directory_urls and posixpath.basename(href) == 'index.html':
            href = posixpath.dirname(href) + '/'
        commit.add_file(git_utils.FileInfo(src, template.render(href=href)))


def list_versions(branch='gh-pages', deploy_prefix=''):
    try:
        return Versions.loads(git_utils.read_file(
            branch, os.path.join(deploy_prefix, versions_file),
            universal_newlines=True
        ))
    except git_utils.GitError:
        return Versions()


def versions_to_file_info(versions, deploy_prefix=''):
    return git_utils.FileInfo(os.path.join(deploy_prefix, versions_file),
                              versions.dumps())


def make_nojekyll():
    return git_utils.FileInfo('.nojekyll', '')


@contextmanager
def deploy(cfg, version, title=None, aliases=[], update_aliases=False,
           alias_type=AliasType.symlink, template=None, *, branch='gh-pages',
           message=None, allow_empty=False, deploy_prefix=''):
    if message is None:
        message = (
            'Deployed {rev} to {doc_version}{deploy_prefix} with MkDocs ' +
            '{mkdocs_version} and mike {mike_version}'
        ).format(
            rev=git_utils.get_latest_commit('HEAD', short=True),
            doc_version=version,
            deploy_prefix=_format_deploy_prefix(deploy_prefix),
            mkdocs_version=mkdocs_utils.version(),
            mike_version=app_version
        )

    all_versions = list_versions(branch, deploy_prefix)
    info = all_versions.add(version, title, aliases, update_aliases)
    version_str = str(info.version)
    destdir = os.path.join(deploy_prefix, version_str)
    alias_destdirs = [os.path.join(deploy_prefix, i) for i in info.aliases]

    # Let the caller perform the build.
    yield

    if alias_type == AliasType.redirect and info.aliases:
        t = _redirect_template(template)

    with git_utils.Commit(branch, message, allow_empty=allow_empty) as commit:
        commit.delete_files([version_str] + list(info.aliases))

        for f in git_utils.walk_real_files(cfg['site_dir']):
            canonical_file = f.copy(destdir, cfg['site_dir'])
            commit.add_file(canonical_file)
            for d in alias_destdirs:
                alias_file = f.copy(d, cfg['site_dir'])
                if alias_type == AliasType.redirect:
                    _add_redirect_to_commit(
                        commit, t, alias_file.path, canonical_file.path,
                        cfg['use_directory_urls']
                    )
                elif alias_type == AliasType.copy:
                    commit.add_file(alias_file)
                elif alias_type != AliasType.symlink:  # pragma: no cover
                    raise ValueError('unrecognized alias type')

        if alias_type == AliasType.symlink:
            for d in alias_destdirs:
                base_dir = os.path.join(d, '..')
                commit.add_file(git_utils.FileInfo(
                    d, os.path.relpath(destdir, base_dir), mode=0o120000
                ))

        commit.add_file(versions_to_file_info(all_versions, deploy_prefix))
        commit.add_file(make_nojekyll())


def delete(identifiers=None, all=False, *, branch='gh-pages', message=None,
           allow_empty=False, deploy_prefix=''):
    if not all and identifiers is None:
        raise ValueError('specify `identifiers` or `all`')

    if message is None:
        message = (
            'Removed {doc_identifiers}{deploy_prefix} with mike {mike_version}'
        ).format(
            doc_identifiers='everything' if all else ', '.join(identifiers),
            deploy_prefix=_format_deploy_prefix(deploy_prefix),
            mike_version=app_version
        )

    with git_utils.Commit(branch, message, allow_empty=allow_empty) as commit:
        if all:
            if deploy_prefix:
                commit.delete_files([deploy_prefix])
            else:
                commit.delete_files('*')
        else:
            all_versions = list_versions(branch, deploy_prefix)
            try:
                removed = all_versions.difference_update(identifiers)
            except KeyError as e:
                raise ValueError('identifier {!r} does not exist'.format(e))

            for i in removed:
                if isinstance(i, str):
                    commit.delete_files([os.path.join(deploy_prefix, i)])
                else:
                    commit.delete_files(
                        [os.path.join(deploy_prefix, str(i.version))] +
                        [os.path.join(deploy_prefix, j) for j in i.aliases]
                    )
            commit.add_file(versions_to_file_info(all_versions, deploy_prefix))


def alias(cfg, identifier, aliases, update_aliases=False,
          alias_type=AliasType.symlink, template=None, *, branch='gh-pages',
          message=None, allow_empty=False, deploy_prefix=''):
    all_versions = list_versions(branch, deploy_prefix)
    try:
        real_version = all_versions.find(identifier, strict=True)[0]
    except KeyError as e:
        raise ValueError('identifier {!r} does not exist'.format(e))

    if message is None:
        message = (
            'Copied {doc_version} to {aliases}{deploy_prefix} with mike ' +
            '{mike_version}'
        ).format(
            doc_version=real_version,
            aliases=', '.join(aliases),
            deploy_prefix=_format_deploy_prefix(deploy_prefix),
            mike_version=app_version
        )

    new_aliases = all_versions.update(real_version, aliases=aliases,
                                      update_aliases=update_aliases)
    destdirs = [os.path.join(deploy_prefix, i) for i in new_aliases]

    if alias_type == AliasType.redirect and destdirs:
        t = _redirect_template(template)

    with git_utils.Commit(branch, message, allow_empty=allow_empty) as commit:
        commit.delete_files(destdirs)

        canonical_dir = os.path.join(deploy_prefix, str(real_version))
        for canonical_file in git_utils.walk_files(branch, canonical_dir):
            for d in destdirs:
                alias_file = canonical_file.copy(d, canonical_dir)
                if alias_type == AliasType.redirect:
                    _add_redirect_to_commit(
                        commit, t, alias_file.path, canonical_file.path,
                        cfg['use_directory_urls']
                    )
                elif alias_type == AliasType.copy:
                    commit.add_file(alias_file)
                elif alias_type != AliasType.symlink:  # pragma: no cover
                    raise ValueError('unrecognized alias type')

        if alias_type == AliasType.symlink:
            for d in destdirs:
                base_dir = os.path.join(d, '..')
                commit.add_file(git_utils.FileInfo(
                    d, os.path.relpath(canonical_dir, base_dir), mode=0o120000
                ))

        commit.add_file(versions_to_file_info(all_versions, deploy_prefix))


def retitle(identifier, title, *, branch='gh-pages', message=None,
            allow_empty=False, deploy_prefix=''):
    if message is None:
        message = (
            'Set title of {doc_identifier} to {title}{deploy_prefix} with ' +
            'mike {mike_version}'
        ).format(
            doc_identifier=identifier,
            title=title,
            deploy_prefix=_format_deploy_prefix(deploy_prefix),
            mike_version=app_version
        )

    all_versions = list_versions(branch, deploy_prefix)
    try:
        all_versions.update(identifier, title)
    except KeyError:
        raise ValueError('identifier {!r} does not exist'.format(identifier))

    with git_utils.Commit(branch, message, allow_empty=allow_empty) as commit:
        commit.add_file(versions_to_file_info(all_versions, deploy_prefix))


def set_default(identifier, template=None, *, branch='gh-pages', message=None,
                allow_empty=False, deploy_prefix=''):
    if message is None:
        message = (
            'Set default version to {doc_identifier}{deploy_prefix} with ' +
            'mike {mike_version}'
        ).format(
            doc_identifier=identifier,
            deploy_prefix=_format_deploy_prefix(deploy_prefix),
            mike_version=app_version
        )

    all_versions = list_versions(branch, deploy_prefix)
    if not all_versions.find(identifier):
        raise ValueError('identifier {!r} does not exist'.format(identifier))

    t = _redirect_template(template)
    with git_utils.Commit(branch, message, allow_empty=allow_empty) as commit:
        commit.add_file(git_utils.FileInfo(
            os.path.join(deploy_prefix, 'index.html'),
            t.render(href=identifier + '/')
        ))


def serve(address='localhost:8000', *, branch='gh-pages', verbose=True):
    my_branch = branch

    class Handler(server.GitBranchHTTPHandler):
        branch = my_branch

    host, *port = address.split(':', 1)
    port = int(port[0]) if port else 8000
    httpd = http.server.HTTPServer((host, port), Handler)

    if verbose:
        print('Starting server at http://{}:{}/'.format(host, port))
        print('Press Ctrl+C to quit.')
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        if verbose:
            print('Stopping server...')
