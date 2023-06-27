# Author: Jakub Andrýsek
# Email: email@kubaandrysek.cz
# Website: https://kubaandrysek.cz
# License: MIT
# GitHub: https://github.com/JakubAndrysek/mkdocs-open-in-new-tab
# PyPI: https://pypi.org/project/mkdocs-open-in-new-tab/
# Inspired by: https://github.com/timvink/mkdocs-charts-plugin/tree/main

import open_in_new_tab
from mkdocs.plugins import BasePlugin
from mkdocs.utils import copy_file
import os

HERE = os.path.dirname(os.path.abspath(__file__))

class OpenInNewTabPlugin(BasePlugin):
    def on_config(self, config, **kwargs):
        """
        Event trigger on config.
        See https://www.mkdocs.org/user-guide/plugins/#on_config.
        """
        # Add pointer to open_in_new_tab.js file to extra_javascript
        # which is added to the output directory during on_post_build() event
        config["extra_javascript"].append("js/open_in_new_tab.js")



    def on_post_build(self, config):
        """
        Event trigger on post build.
        See https://www.mkdocs.org/user-guide/plugins/#on_post_build.
        """

        js_output_base_path = os.path.join(config["site_dir"], "js")
        js_file_path = os.path.join(js_output_base_path, "open_in_new_tab.js")
        package = os.path.dirname(os.path.abspath(__file__))
        copy_file(
            os.path.join(os.path.join(package, "js"), "open_in_new_tab.js"),
            js_file_path,
        )