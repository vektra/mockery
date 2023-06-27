import mkdocs.config
import mkdocs.utils
import os
import re
import subprocess
import yaml
from collections.abc import Iterable, Mapping
from contextlib import contextmanager
from tempfile import NamedTemporaryFile

docs_version_var = 'MIKE_DOCS_VERSION'


def _open_config(config_file=None):
    if config_file is None:
        config_file = ['mkdocs.yml', 'mkdocs.yaml']
    elif not isinstance(config_file, Iterable) or isinstance(config_file, str):
        config_file = [config_file]

    exc = None
    for i in config_file:
        try:
            return open(i, 'rb')
        except FileNotFoundError as e:
            if not exc:
                exc = e
    raise exc


def load_config(config_file=None, **kwargs):
    with _open_config(config_file) as f:
        cfg = mkdocs.config.load_config(f, **kwargs)
        return cfg['plugins'].run_event('config', cfg)


@contextmanager
def inject_plugin(config_file):
    with _open_config(config_file) as f:
        config_file = f.name
        config = mkdocs.utils.yaml_load(f)

    plugins = config.setdefault('plugins', ['search'])
    for i in plugins:
        if ( (isinstance(i, str) and i == 'mike') or
             (isinstance(i, dict) and 'mike' in i) ):
            yield config_file
            return

    if isinstance(plugins, Mapping):
        config['plugins'] = {'mike': {}}
        config['plugins'].update(plugins)
    else:
        plugins.insert(0, 'mike')

    with NamedTemporaryFile(mode='w', dir=os.path.dirname(config_file),
                            prefix='mike-mkdocs', suffix='.yml',
                            delete=False) as f:
        yaml.dump(config, f, sort_keys=False)

    try:
        yield f.name
    finally:
        os.remove(f.name)


def build(config_file, version, verbose=True):
    command = (
        ['mkdocs', 'build', '--clean'] +
        (['--config-file', config_file] if config_file else [])
    )

    env = os.environ.copy()
    env[docs_version_var] = version

    output = None if verbose else subprocess.DEVNULL
    subprocess.run(command, check=True, env=env, stdout=output, stderr=output)


def version():
    output = subprocess.run(
        ['mkdocs', '--version'],
        check=True, stdout=subprocess.PIPE, universal_newlines=True
    ).stdout.rstrip()
    m = re.search('^mkdocs, version (\\S*)', output)
    return m.group(1)
