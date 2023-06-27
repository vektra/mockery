# Author: Jakub Andrýsek
# Email: email@kubaandrysek.cz
# Website: https://kubaandrysek.cz
# License: MIT
# GitHub: https://github.com/JakubAndrysek/mkdocs-open-in-new-tab

class OpenInNewTabPlugin(BasePlugin):
    def on_page_markdown(self, markdown, **kwargs):
        return markdown_processor(markdown)

    def on_page_content(self, html, **kwargs):
        return html_processor(html)

    def on_post_build(self, config):
        # add js file to extra_javascript (js file is located in ../js/open_in_new_tab.js)
        js_path = os.path.join(os.path.dirname(__file__), 'js', 'open_in_new_tab.js')
        config['extra_javascript'].append(js_path)