import re
from typing import Iterator, List, Tuple

from .baseversion import *
from .basespecifier import *

__all__ = ["InvalidVersion", "InvalidSpecifier", "LooseSpecifier",
           "LooseSpecifierSet", "LooseVersion"]

LooseCmpKey = Tuple[str, ...]


class LooseVersion(BaseVersion):
    def __init__(self, version: str) -> None:
        self._version = str(version)
        self._key = _loose_cmpkey(self._version)

    def __str__(self) -> str:
        return self._version

    @property
    def public(self) -> str:
        return self._version

    @property
    def base_version(self) -> str:
        return self._version

    @property
    def epoch(self) -> int:
        return 0

    @property
    def release(self) -> None:
        return None

    @property
    def pre(self) -> None:
        return None

    @property
    def post(self) -> None:
        return None

    @property
    def dev(self) -> None:
        return None

    @property
    def local(self) -> None:
        return None

    @property
    def is_prerelease(self) -> bool:
        return False

    @property
    def is_postrelease(self) -> bool:
        return False

    @property
    def is_devrelease(self) -> bool:
        return False


_loose_version_component_re = re.compile(r"(\d+ | [a-z]+ | \.| -)", re.VERBOSE)

_loose_version_replacement_map = {
    "pre": "c",
    "preview": "c",
    "-": "final-",
    "rc": "c",
    "dev": "@",
}


def _parse_version_parts(s: str) -> Iterator[str]:
    for part in _loose_version_component_re.split(s):
        part = _loose_version_replacement_map.get(part, part)

        if not part or part == ".":
            continue

        if part[:1] in "0123456789":
            # pad for numeric comparison
            yield part.zfill(8)
        else:
            yield "*" + part

    # ensure that alpha/beta/candidate are before final
    yield "*final"


def _loose_cmpkey(version: str) -> LooseCmpKey:
    # This scheme is taken from pkg_resources.parse_version setuptools prior to
    # it's adoption of the packaging library.
    parts: List[str] = []
    for part in _parse_version_parts(version.lower()):
        if part.startswith("*"):
            # remove "-" before a prerelease tag
            if part < "*final":
                while parts and parts[-1] == "*final-":
                    parts.pop()

            # remove trailing zeros from each series of numeric parts
            while parts and parts[-1] == "00000000":
                parts.pop()

        parts.append(part)

    return tuple(parts)


class LooseSpecifier(IndividualSpecifier):
    _regex_str = r"""
        (?P<operator>(==|!=|<=|>=|<|>))
        \s*
        (?P<version>
            [^,;\s)]* # Since this is a "loose" specifier, and the version
                      # string can be just about anything, we match everything
                      # except for whitespace, a semi-colon for marker support,
                      # a closing paren since versions can be enclosed in
                      # them, and a comma since it's a version separator.
        )
        """

    _regex = re.compile(r"^\s*" + _regex_str + r"\s*$",
                        re.VERBOSE | re.IGNORECASE)

    _operators = {
        "==": "equal",
        "!=": "not_equal",
        "<=": "less_than_equal",
        ">=": "greater_than_equal",
        "<": "less_than",
        ">": "greater_than",
    }

    def _coerce_version(self, version: UnparsedVersion) -> LooseVersion:
        if not isinstance(version, LooseVersion):
            version = LooseVersion(str(version))
        return version

    @property
    def _canonical_spec(self) -> Tuple[str, str]:
        return self._spec

    def _compare_equal(self, prospective: LooseVersion, spec: str) -> bool:
        return prospective == self._coerce_version(spec)

    def _compare_not_equal(self, prospective: LooseVersion, spec: str) -> bool:
        return prospective != self._coerce_version(spec)

    def _compare_less_than_equal(self, prospective: LooseVersion,
                                 spec: str) -> bool:
        return prospective <= self._coerce_version(spec)

    def _compare_greater_than_equal(self, prospective: LooseVersion,
                                    spec: str) -> bool:
        return prospective >= self._coerce_version(spec)

    def _compare_less_than(self, prospective: LooseVersion, spec: str) -> bool:
        return prospective < self._coerce_version(spec)

    def _compare_greater_than(self, prospective: LooseVersion,
                              spec: str) -> bool:
        return prospective > self._coerce_version(spec)


class LooseSpecifierSet(BaseSpecifierSet):
    def __init__(self, specifiers: str = "") -> None:
        # Split on , to break each individual specifier into its own item, and
        # strip each item to remove leading/trailing whitespace.
        split_specifiers = [s.strip() for s in specifiers.split(",")
                            if s.strip()]

        # Parse each individual specifier as a LooseSpecifier.
        parsed: Set[BaseSpecifier] = set(LooseSpecifier(specifier)
                                         for specifier in split_specifiers)

        super().__init__(parsed, None)

    def _coerce_version(self, version: UnparsedVersion) -> LooseVersion:
        if not isinstance(version, LooseVersion):
            version = LooseVersion(str(version))
        return version

    def _filter_prereleases(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool]
    ) -> Iterable[UnparsedVersion]:
        # Note: We ignore prereleases, since LooseVersions are never
        # prereleases, and only have that field for compatibility.
        return iterable


Version = LooseVersion
Specifier = LooseSpecifier
SpecifierSet = LooseSpecifierSet
