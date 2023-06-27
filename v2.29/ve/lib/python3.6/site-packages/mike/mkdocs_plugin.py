import importlib_metadata as metadata
import os
from urllib.parse import urljoin
from mkdocs.config import config_options as opts
from mkdocs.plugins import BasePlugin
from mkdocs.structure.files import File

from .mkdocs_utils import docs_version_var
from .commands import AliasType

try:
    from mkdocs.exceptions import PluginError
except ImportError:  # pragma: no cover
    PluginError = ValueError


def get_theme_dir(theme_name):
    try:
        theme = metadata.entry_points(group='mike.themes')[theme_name]
    except KeyError:
        raise ValueError("theme '{}' unsupported".format(theme_name))
    return os.path.dirname(theme.load().__file__)


class MikePlugin(BasePlugin):
    config_scheme = (
        ('alias_type', opts.Choice(tuple(i.name for i in AliasType),
                                   default='symlink')),
        ('redirect_template', opts.Type((str, type(None)), default=None)),
        ('deploy_prefix', opts.Type(str, default='')),
        ('version_selector', opts.Type(bool, default=True)),
        ('canonical_version', opts.Type((str, type(None)), default=None)),
        ('css_dir', opts.Type(str, default='css')),
        ('javascript_dir', opts.Type(str, default='js')),
    )

    @classmethod
    def default(cls):
        plugin = cls()
        plugin.load_config({})
        plugin.on_config({})
        return plugin

    def on_config(self, config):
        version = os.environ.get(docs_version_var)
        if version and config.get('site_url'):
            if self.config['canonical_version'] is not None:
                version = self.config['canonical_version']
            config['site_url'] = urljoin(config['site_url'], version)

    def on_files(self, files, config):
        if not self.config['version_selector']:
            return files

        try:
            theme_dir = get_theme_dir(config['theme'].name)
        except ValueError:
            return files

        for path, prop in [('css', 'css'), ('js', 'javascript')]:
            cfg_value = self.config[prop + '_dir']
            srcdir = os.path.join(theme_dir, path)
            destdir = os.path.join(config['site_dir'], cfg_value)

            extra_kind = 'extra_' + prop
            norm_extras = [os.path.normpath(i) for i in config[extra_kind]]
            for f in os.listdir(srcdir):
                relative_dest = os.path.join(cfg_value, f)
                if relative_dest in norm_extras:
                    raise PluginError('{!r} is already included in {!r}'
                                      .format(relative_dest, extra_kind))

                files.append(File(f, srcdir, destdir, False))
                config[extra_kind].append(relative_dest)
        return files
