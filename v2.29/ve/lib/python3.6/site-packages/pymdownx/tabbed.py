"""
Tabbed.

pymdownx.tabbed

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
from markdown.blockprocessors import BlockProcessor
import xml.etree.ElementTree as etree
import re


class TabbedProcessor(BlockProcessor):
    """Tabbed block processor."""

    START = re.compile(
        r'(?:^|\n)={3}(!)? +"(.*?)" *(?:\n|$)'
    )
    COMPRESS_SPACES = re.compile(r' {2,}')

    def __init__(self, *args, alternate_style=False):
        """Initialize."""

        super().__init__(*args)
        self.tab_group_count = 0
        self.current_sibling = None
        self.content_indention = 0
        self.alternate_style = alternate_style

    def detab_by_length(self, text, length):
        """Remove a tab from the front of each line of the given text."""

        newtext = []
        lines = text.split('\n')
        for line in lines:
            if line.startswith(' ' * length):
                newtext.append(line[length:])
            elif not line.strip():
                newtext.append('')  # pragma: no cover
            else:
                break
        return '\n'.join(newtext), '\n'.join(lines[len(newtext):])

    def parse_content(self, parent, block):
        """Get sibling tab.

        Retrieve the appropriate sibling element. This can get tricky when
        dealing with lists.

        """

        old_block = block
        non_tabs = ''
        tabbed_set = 'tabbed-set' if not self.alternate_style else 'tabbed-set tabbed-alternate'

        # We already acquired the block via test
        if self.current_sibling is not None:
            sibling = self.current_sibling
            block, non_tabs = self.detab_by_length(block, self.content_indent)
            self.current_sibling = None
            self.content_indent = 0
            return sibling, block, non_tabs

        sibling = self.lastChild(parent)

        if sibling is None or sibling.tag.lower() != 'div' or sibling.attrib.get('class', '') != tabbed_set:
            sibling = None
        else:
            # If the last child is a list and the content is indented sufficient
            # to be under it, then the content's is sibling is in the list.
            if self.alternate_style:
                last_child = self.lastChild(self.lastChild(sibling))
                tabbed_content = 'tabbed-block'
            else:
                last_child = self.lastChild(sibling)
                tabbed_content = 'tabbed-content'
            child_class = last_child.attrib.get('class', '') if last_child else ''
            indent = 0
            while last_child:
                if (
                    sibling and block.startswith(' ' * self.tab_length * 2) and
                    last_child and (
                        last_child.tag in ('ul', 'ol', 'dl') or
                        (
                            last_child.tag == 'div' and
                            child_class == tabbed_content
                        )
                    )
                ):

                    # Handle nested tabbed content
                    if last_child.tag == 'div' and child_class == tabbed_content:
                        temp_child = self.lastChild(last_child)
                        if temp_child.tag not in ('ul', 'ol', 'dl'):
                            break
                        last_child = temp_child
                        child_class = last_child.attrib.get('class', '') if last_child else ''

                    # The expectation is that we'll find an `<li>`.
                    # We should get it's last child as well.
                    sibling = self.lastChild(last_child)
                    last_child = self.lastChild(sibling) if sibling else None
                    child_class = last_child.attrib.get('class', '') if last_child else ''

                    # Context has been lost at this point, so we must adjust the
                    # text's indentation level so it will be evaluated correctly
                    # under the list.
                    block = block[self.tab_length:]
                    indent += self.tab_length
                else:
                    last_child = None

            if not block.startswith(' ' * self.tab_length):
                sibling = None

            if sibling is not None:
                indent += self.tab_length
                block, non_tabs = self.detab_by_length(old_block, indent)
                self.current_sibling = sibling
                self.content_indent = indent

        return sibling, block, non_tabs

    def test(self, parent, block):
        """Test block."""

        if self.START.search(block):
            return True
        else:
            return self.parse_content(parent, block)[0] is not None

    def run(self, parent, blocks):
        """Convert to tabbed block."""

        block = blocks.pop(0)
        m = self.START.search(block)
        tabbed_set = 'tabbed-set' if not self.alternate_style else 'tabbed-set tabbed-alternate'

        if m:
            # removes the first line
            if m.start() > 0:
                self.parser.parseBlocks(parent, [block[:m.start()]])
            block = block[m.end():]
            sibling = self.lastChild(parent)
            block, non_tabs = self.detab(block)
        else:
            sibling, block, non_tabs = self.parse_content(parent, block)

        if m:
            special = m.group(1) if m.group(1) else ''
            title = m.group(2) if m.group(2) else m.group(3)

            if (
                sibling and sibling.tag.lower() == 'div' and
                sibling.attrib.get('class', '') == tabbed_set and
                special != '!'
            ):
                first = False
                sfences = sibling
                if self.alternate_style:
                    index = [index for index, last_input in enumerate(sfences.findall('input'), 1)][-1]
                    labels = None
                    content = None
                    for d in sfences.findall('div'):
                        if d.attrib['class'] == 'tabbed-labels':
                            labels = d
                        elif d.attrib['class'] == 'tabbed-content':
                            content = d
                        if labels is not None and content is not None:
                            break
            else:
                first = True
                self.tab_group_count += 1
                sfences = etree.SubElement(
                    parent,
                    'div',
                    {'class': tabbed_set, 'data-tabs': '%d:0' % self.tab_group_count}
                )
                if self.alternate_style:
                    index = 0
                    labels = etree.SubElement(
                        sfences,
                        'div',
                        {'class': 'tabbed-labels'}
                    )
                    content = etree.SubElement(
                        sfences,
                        'div',
                        {'class': 'tabbed-content'}
                    )

            data = sfences.attrib['data-tabs'].split(':')
            tab_set = int(data[0])
            tab_count = int(data[1]) + 1

            attributes = {
                "name": "__tabbed_%d" % tab_set,
                "type": "radio",
                "id": "__tabbed_%d_%d" % (tab_set, tab_count)
            }

            if first:
                attributes['checked'] = 'checked'

            if self.alternate_style:
                input_el = etree.Element(
                    'input',
                    attributes
                )
                sfences.insert(index, input_el)
                lab = etree.SubElement(
                    labels,
                    "label",
                    {
                        "for": "__tabbed_%d_%d" % (tab_set, tab_count)
                    }
                )
                lab.text = title

                div = etree.SubElement(
                    content,
                    "div",
                    {'class': 'tabbed-block'}
                )
            else:
                etree.SubElement(
                    sfences,
                    'input',
                    attributes
                )
                lab = etree.SubElement(
                    sfences,
                    "label",
                    {
                        "for": "__tabbed_%d_%d" % (tab_set, tab_count)
                    }
                )
                lab.text = title

                div = etree.SubElement(
                    sfences,
                    "div",
                    {
                        "class": "tabbed-content"
                    }
                )

            sfences.attrib['data-tabs'] = '%d:%d' % (tab_set, tab_count)
        else:
            if sibling.tag in ('li', 'dd') and sibling.text:
                # Sibling is a list item, but we need to wrap it's content should be wrapped in <p>
                text = sibling.text
                sibling.text = ''
                p = etree.SubElement(sibling, 'p')
                p.text = text
                div = sibling
            elif sibling.tag == 'div' and sibling.attrib.get('class', '') == tabbed_set:
                # Get `tabbed-content` under `tabbed-set`
                if self.alternate_style:
                    div = self.lastChild(self.lastChild(sibling))
                else:
                    div = self.lastChild(sibling)
            else:
                # Pass anything else as the parent
                div = sibling

        self.parser.parseChunk(div, block)

        if non_tabs:
            # Insert the tabbed content back into blocks
            blocks.insert(0, non_tabs)


class TabbedExtension(Extension):
    """Add Tabbed extension."""

    def __init__(self, *args, **kwargs):
        """Initialize."""

        self.config = {
            'alternate_style': [False, "Use alternate style - Default: False"]
        }

        super(TabbedExtension, self).__init__(*args, **kwargs)

    def extendMarkdown(self, md):
        """Add Tabbed to Markdown instance."""
        md.registerExtension(self)

        config = self.getConfigs()
        self.tab_processor = TabbedProcessor(md.parser, alternate_style=config['alternate_style'])
        md.parser.blockprocessors.register(self.tab_processor, "tabbed", 105)

    def reset(self):
        """Reset."""

        self.tab_processor.tab_group_count = 0


def makeExtension(*args, **kwargs):
    """Return extension."""

    return TabbedExtension(*args, **kwargs)
