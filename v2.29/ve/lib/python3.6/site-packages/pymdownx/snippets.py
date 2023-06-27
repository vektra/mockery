"""
Snippet ---8<---.

pymdownx.snippet
Inject snippets

MIT license.

Copyright (c) 2017 Isaac Muse <isaacmuse@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
documentation files (the "Software"), to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions
of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
DEALINGS IN THE SOFTWARE.
"""
from markdown import Extension
from markdown.preprocessors import Preprocessor
import re
import codecs
import os


class SnippetPreprocessor(Preprocessor):
    """Handle snippets in Markdown content."""

    RE_ALL_SNIPPETS = re.compile(
        r'''(?x)
        ^(?P<space>[ \t]*)
        (?P<all>
            (?P<inline_marker>-{2,}8<-{2,}[ \t]+)
            (?P<snippet>(?:"(?:\\"|[^"\n\r])+?"|'(?:\\'|[^'\n\r])+?'))(?![ \t]) |
            (?P<block_marker>-{2,}8<-{2,})(?![ \t])
        )\r?$
        '''
    )

    RE_SNIPPET = re.compile(
        r'''(?x)
        ^(?P<space>[ \t]*)
        (?P<snippet>.*?)\r?$
        '''
    )

    def __init__(self, config, md):
        """Initialize."""

        base = config.get('base_path')
        if isinstance(base, str):
            base = [base]
        self.base_path = base
        self.encoding = config.get('encoding')
        self.check_paths = config.get('check_paths')
        self.auto_append = config.get('auto_append')
        self.tab_length = md.tab_length
        super(SnippetPreprocessor, self).__init__()

    def get_snippet(self, path):
        """Get snippet."""

        snippet = None
        for base in self.base_path:
            if os.path.exists(base):
                if os.path.isdir(base):
                    filename = os.path.join(base, path)
                    if os.path.exists(filename):
                        snippet = filename
                        break
                else:
                    basename = os.path.basename(base)
                    dirname = os.path.dirname(base)
                    if basename.lower() == path.lower():
                        filename = os.path.join(dirname, path)
                        if os.path.exists(filename):
                            snippet = filename
                            break
        return snippet

    def parse_snippets(self, lines, file_name=None):
        """Parse snippets snippet."""

        new_lines = []
        inline = False
        block = False
        for line in lines:
            inline = False
            m = self.RE_ALL_SNIPPETS.match(line)
            if m:
                if block and m.group('inline_marker'):
                    # Don't use inline notation directly under a block.
                    # It's okay if inline is used again in sub file though.
                    continue
                elif m.group('inline_marker'):
                    # Inline
                    inline = True
                else:
                    # Block
                    block = not block
                    continue
            elif not block:
                # Not in snippet, and we didn't find an inline,
                # so just a normal line
                new_lines.append(line)
                continue

            if block and not inline:
                # We are in a block and we didn't just find a nested inline
                # So check if a block path
                m = self.RE_SNIPPET.match(line)

            if m:
                # Get spaces and snippet path.  Remove quotes if inline.
                space = m.group('space').expandtabs(self.tab_length)
                path = m.group('snippet')[1:-1].strip() if inline else m.group('snippet').strip()

                if not inline:
                    # Block path handling
                    if not path:
                        # Empty path line, insert a blank line
                        new_lines.append('')
                        continue
                if path.startswith('; '):
                    # path stats with '#', consider it commented out.
                    # We just removing the line.
                    continue

                snippet = self.get_snippet(path)
                if snippet:
                    if snippet in self.seen:
                        # This is in the stack and we don't want an infinite loop!
                        continue
                    if file_name:
                        # Track this file.
                        self.seen.add(file_name)
                    try:
                        with codecs.open(snippet, 'r', encoding=self.encoding) as f:
                            new_lines.extend(
                                [space + l2 for l2 in self.parse_snippets([l.rstrip('\r\n') for l in f], snippet)]
                            )
                    except Exception:  # pragma: no cover
                        pass
                    if file_name:
                        self.seen.remove(file_name)
                elif self.check_paths:
                    raise IOError("Snippet at path %s could not be found" % path)

        return new_lines

    def run(self, lines):
        """Process snippets."""

        self.seen = set()
        if self.auto_append:
            lines.extend("\n\n--8<--\n{}\n--8<--\n".format('\n\n'.join(self.auto_append)).split('\n'))

        return self.parse_snippets(lines)


class SnippetExtension(Extension):
    """Snippet extension."""

    def __init__(self, *args, **kwargs):
        """Initialize."""

        self.config = {
            'base_path': [["."], "Base path for snippet paths - Default: [\".\"]"],
            'encoding': ["utf-8", "Encoding of snippets - Default: \"utf-8\""],
            'check_paths': [False, "Make the build fail if a snippet can't be found - Default: \"false\""],
            "auto_append": [
                [],
                "A list of snippets (relative to the 'base_path') to auto append to the Markdown content - Default: []"
            ]
        }

        super(SnippetExtension, self).__init__(*args, **kwargs)

    def extendMarkdown(self, md):
        """Register the extension."""

        self.md = md
        md.registerExtension(self)
        config = self.getConfigs()
        snippet = SnippetPreprocessor(config, md)
        md.preprocessors.register(snippet, "snippet", 32)


def makeExtension(*args, **kwargs):
    """Return extension."""

    return SnippetExtension(*args, **kwargs)
