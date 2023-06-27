import itertools
import re
from typing import List, NamedTuple, Optional, SupportsInt, Tuple

from .baseversion import *
from .basespecifier import *
from .infinity import *

__all__ = ["InvalidVersion", "InvalidSpecifier", "PythonSpecifier",
           "PythonSpecifierSet", "PythonVersion"]

LocalType = Tuple[Union[int, str], ...]
PrePostDevType = Union[InfiniteTypes, LetterVersion]
LocalCmpType = Union[
    NegativeInfinityType,
    Tuple[
        Union[
            Tuple[int, str],
            Tuple[InfiniteTypes, str],
        ], ...
    ],
]
PythonCmpKey = Tuple[int, Tuple[int, ...], PrePostDevType, PrePostDevType,
                     PrePostDevType, LocalCmpType]


class _Version(NamedTuple):
    epoch: int
    release: Tuple[int, ...]
    dev: Optional[LetterVersion]
    pre: Optional[LetterVersion]
    post: Optional[LetterVersion]
    local: Optional[LocalType]


# Deliberately not anchored to the start and end of the string, to make it
# easier for 3rd party code to reuse
VERSION_PATTERN = r"""
    v?
    (?:
        (?:(?P<epoch>[0-9]+)!)?                           # epoch
        (?P<release>[0-9]+(?:\.[0-9]+)*)                  # release segment
        (?P<pre>                                          # pre-release
            [-_\.]?
            (?P<pre_l>(a|b|c|rc|alpha|beta|pre|preview))
            [-_\.]?
            (?P<pre_n>[0-9]+)?
        )?
        (?P<post>                                         # post release
            (?:-(?P<post_n1>[0-9]+))
            |
            (?:
                [-_\.]?
                (?P<post_l>post|rev|r)
                [-_\.]?
                (?P<post_n2>[0-9]+)?
            )
        )?
        (?P<dev>                                          # dev release
            [-_\.]?
            (?P<dev_l>dev)
            [-_\.]?
            (?P<dev_n>[0-9]+)?
        )?
    )
    (?:\+(?P<local>[a-z0-9]+(?:[-_\.][a-z0-9]+)*))?       # local version
"""


class PythonVersion(BaseVersion):
    _regex = re.compile(r"^\s*" + VERSION_PATTERN + r"\s*$",
                        re.VERBOSE | re.IGNORECASE)

    def __init__(self, version: str) -> None:
        # Validate the version and parse it into pieces
        match = self._regex.search(version)
        if not match:
            raise InvalidVersion("Invalid version: '{0}'".format(version))

        # Store the parsed out pieces of the version
        self._version = _Version(
            epoch=int(match.group("epoch")) if match.group("epoch") else 0,
            release=tuple(int(i) for i in match.group("release").split(".")),
            pre=_parse_letter_version(match.group("pre_l"),
                                      match.group("pre_n")),
            post=_parse_letter_version(
                match.group("post_l"),
                match.group("post_n1") or match.group("post_n2")
            ),
            dev=_parse_letter_version(match.group("dev_l"),
                                      match.group("dev_n")),
            local=_parse_local_version(match.group("local")),
        )

        # Generate a key which will be used for sorting
        self._key = _cmpkey(
            self._version.epoch,
            self._version.release,
            self._version.pre,
            self._version.post,
            self._version.dev,
            self._version.local,
        )

    def __str__(self) -> str:
        parts = []

        # Epoch
        if self.epoch != 0:
            parts.append("{0}!".format(self.epoch))

        # Release segment
        parts.append(".".join(str(x) for x in self.release))

        # Pre-release
        if self.pre is not None:
            parts.append("".join(str(x) for x in self.pre))

        # Post-release
        if self.post is not None:
            parts.append(".post{0}".format(self.post))

        # Development release
        if self.dev is not None:
            parts.append(".dev{0}".format(self.dev))

        # Local version segment
        if self.local is not None:
            parts.append("+{0}".format(self.local))

        return "".join(parts)

    @property
    def epoch(self) -> int:
        return self._version.epoch

    @property
    def release(self) -> Tuple[int, ...]:
        return self._version.release

    @property
    def pre(self) -> Optional[LetterVersion]:
        return self._version.pre

    @property
    def post(self) -> Optional[int]:
        return self._version.post[1] if self._version.post else None

    @property
    def dev(self) -> Optional[int]:
        return self._version.dev[1] if self._version.dev else None

    @property
    def local(self) -> Optional[str]:
        if self._version.local:
            return ".".join(str(x) for x in self._version.local)
        else:
            return None

    @property
    def public(self) -> str:
        return str(self).split("+", 1)[0]

    @property
    def base_version(self) -> str:
        parts = []

        # Epoch
        if self.epoch != 0:
            parts.append("{0}!".format(self.epoch))

        # Release segment
        parts.append(".".join(str(x) for x in self.release))

        return "".join(parts)

    @property
    def is_prerelease(self) -> bool:
        return self.dev is not None or self.pre is not None

    @property
    def is_postrelease(self) -> bool:
        return self.post is not None

    @property
    def is_devrelease(self) -> bool:
        return self.dev is not None

    @property
    def major(self) -> int:
        return self.release[0] if len(self.release) >= 1 else 0

    @property
    def minor(self) -> int:
        return self.release[1] if len(self.release) >= 2 else 0

    @property
    def micro(self) -> int:
        return self.release[2] if len(self.release) >= 3 else 0


def _parse_letter_version(
    letter: str, number: Union[str, bytes, SupportsInt]
) -> Optional[LetterVersion]:
    if letter:
        # We consider there to be an implicit 0 in a pre-release if there is
        # not a numeral associated with it.
        if number is None:
            number = 0

        # We normalize any letters to their lower case form
        letter = letter.lower()

        # We consider some words to be alternate spellings of other words and
        # in those cases we want to normalize the spellings to our preferred
        # spelling.
        if letter == "alpha":
            letter = "a"
        elif letter == "beta":
            letter = "b"
        elif letter in ["c", "pre", "preview"]:
            letter = "rc"
        elif letter in ["rev", "r"]:
            letter = "post"

        return letter, int(number)
    if not letter and number:
        # We assume if we are given a number, but we are not given a letter
        # then this is using the implicit post release syntax (e.g. 1.0-1)
        letter = "post"

        return letter, int(number)

    return None


_local_version_separators = re.compile(r"[\._-]")


def _parse_local_version(local: str) -> Optional[LocalType]:
    """
    Takes a string like abc.1.twelve and turns it into ("abc", 1, "twelve").
    """
    if local is not None:
        return tuple(
            part.lower() if not part.isdigit() else int(part)
            for part in _local_version_separators.split(local)
        )
    return None


def _cmpkey(
    epoch: int,
    release: Tuple[int, ...],
    pre: Optional[LetterVersion],
    post: Optional[LetterVersion],
    dev: Optional[LetterVersion],
    local: Optional[LocalType],
) -> PythonCmpKey:
    # When we compare a release version, we want to compare it with all of the
    # trailing zeros removed. So we'll use a reverse the list, drop all the now
    # leading zeros until we come to something non zero, then take the rest
    # re-reverse it back into the correct order and make it a tuple and use
    # that for our sorting key.
    _release = tuple(reversed(list(itertools.dropwhile(
        lambda x: x == 0, reversed(release)
    ))))

    # We need to "trick" the sorting algorithm to put 1.0.dev0 before 1.0a0.
    # We'll do this by abusing the pre segment, but we _only_ want to do this
    # if there is not a pre or a post segment. If we have one of those then
    # the normal sorting rules will handle this case correctly.
    if pre is None and post is None and dev is not None:
        _pre: PrePostDevType = NegativeInfinity
    # Versions without a pre-release (except as noted above) should sort after
    # those with one.
    elif pre is None:
        _pre = Infinity
    else:
        _pre = pre

    # Versions without a post segment should sort before those with one.
    if post is None:
        _post: PrePostDevType = NegativeInfinity
    else:
        _post = post

    # Versions without a development segment should sort after those with one.
    if dev is None:
        _dev: PrePostDevType = Infinity
    else:
        _dev = dev

    if local is None:
        # Versions without a local segment should sort before those with one.
        _local: LocalCmpType = NegativeInfinity
    else:
        # Versions with a local segment need that segment parsed to implement
        # the sorting rules in PEP440.
        # - Alpha numeric segments sort before numeric segments
        # - Alpha numeric segments sort lexicographically
        # - Numeric segments sort numerically
        # - Shorter versions sort before longer versions when the prefixes
        #   match exactly
        _local = tuple(
            (i, "") if isinstance(i, int) else (NegativeInfinity, i)
            for i in local
        )

    return epoch, _release, _pre, _post, _dev, _local


class PythonSpecifier(IndividualSpecifier):
    _regex_str = r"""
        (?P<operator>(~=|==|!=|<=|>=|<|>|===))
        (?P<version>
            (?:
                # The identity operators allow for an escape hatch that will
                # do an exact string match of the version you wish to install.
                # This will not be parsed by PEP 440 and we cannot determine
                # any semantic meaning from it. This operator is discouraged
                # but included entirely as an escape hatch.
                (?<====)  # Only match for the identity operator
                \s*
                [^\s]*    # We just match everything, except for whitespace
                          # since we are only testing for strict identity.
            )
            |
            (?:
                # The (non)equality operators allow for wild card and local
                # versions to be specified so we have to define these two
                # operators separately to enable that.
                (?<===|!=)            # Only match for equals and not equals

                \s*
                v?
                (?:[0-9]+!)?          # epoch
                [0-9]+(?:\.[0-9]+)*   # release
                (?:                   # pre release
                    [-_\.]?
                    (a|b|c|rc|alpha|beta|pre|preview)
                    [-_\.]?
                    [0-9]*
                )?
                (?:                   # post release
                    (?:-[0-9]+)|(?:[-_\.]?(post|rev|r)[-_\.]?[0-9]*)
                )?

                # You cannot use a wild card and a dev or local version
                # together so group them with a | and make them optional.
                (?:
                    (?:[-_\.]?dev[-_\.]?[0-9]*)?         # dev release
                    (?:\+[a-z0-9]+(?:[-_\.][a-z0-9]+)*)? # local
                    |
                    \.\*  # Wild card syntax of .*
                )?
            )
            |
            (?:
                # The compatible operator requires at least two digits in the
                # release segment.
                (?<=~=)               # Only match for the compatible operator

                \s*
                v?
                (?:[0-9]+!)?          # epoch
                [0-9]+(?:\.[0-9]+)+   # release  (We have a + instead of a *)
                (?:                   # pre release
                    [-_\.]?
                    (a|b|c|rc|alpha|beta|pre|preview)
                    [-_\.]?
                    [0-9]*
                )?
                (?:                                   # post release
                    (?:-[0-9]+)|(?:[-_\.]?(post|rev|r)[-_\.]?[0-9]*)
                )?
                (?:[-_\.]?dev[-_\.]?[0-9]*)?          # dev release
            )
            |
            (?:
                # All other operators only allow a sub set of what the
                # (non)equality operators do. Specifically they do not allow
                # local versions to be specified nor do they allow the prefix
                # matching wild cards.
                (?<!==|!=|~=)         # We have special cases for these
                                      # operators so we want to make sure they
                                      # don't match here.

                \s*
                v?
                (?:[0-9]+!)?          # epoch
                [0-9]+(?:\.[0-9]+)*   # release
                (?:                   # pre release
                    [-_\.]?
                    (a|b|c|rc|alpha|beta|pre|preview)
                    [-_\.]?
                    [0-9]*
                )?
                (?:                                   # post release
                    (?:-[0-9]+)|(?:[-_\.]?(post|rev|r)[-_\.]?[0-9]*)
                )?
                (?:[-_\.]?dev[-_\.]?[0-9]*)?          # dev release
            )
        )
        """

    _regex = re.compile(r"^\s*" + _regex_str + r"\s*$",
                        re.VERBOSE | re.IGNORECASE)

    _operators = {
        "~=": "compatible",
        "==": "equal",
        "!=": "not_equal",
        "<=": "less_than_equal",
        ">=": "greater_than_equal",
        "<": "less_than",
        ">": "greater_than",
        "===": "arbitrary",
    }

    def _coerce_version(self, version: UnparsedVersion) -> PythonVersion:
        if not isinstance(version, PythonVersion):
            version = PythonVersion(str(version))
        return version

    @property
    def _canonical_spec(self) -> Tuple[str, str]:
        return self._spec[0], _canonicalize_version(self._spec[1])

    def _compare_compatible(self, prospective: BaseVersion, spec: str) -> bool:
        # Compatible releases have an equivalent combination of >= and ==. That
        # is that ~=2.2 is equivalent to >=2.2,==2.*. This allows us to
        # implement this in terms of the other specifiers instead of
        # implementing it ourselves. The only thing we need to do is construct
        # the other specifiers.

        # We want everything but the last item in the version, but we want to
        # ignore post and dev releases and we want to treat the pre-release as
        # it's own separate segment.
        prefix = ".".join(list(
            itertools.takewhile(
                lambda x: (not x.startswith("post") and not
                           x.startswith("dev")),
                _version_split(spec),
            )
        )[:-1])

        # Add the prefix notation to the end of our string
        prefix += ".*"

        return (self._get_operator(">=")(prospective, spec) and
                self._get_operator("==")(prospective, prefix))

    def _compare_equal(self, prospective: BaseVersion, spec: str) -> bool:
        # We need special logic to handle prefix matching
        if spec.endswith(".*"):
            # In the case of prefix matching we want to ignore local segment.
            prospective = PythonVersion(prospective.public)
            # Split the spec out by dots, and pretend that there is an implicit
            # dot in between a release segment and a pre-release segment.
            split_spec = _version_split(spec[:-2])  # Remove the trailing .*

            # Split the prospective version out by dots, and pretend that there
            # is an implicit dot in between a release segment and a pre-release
            # segment.
            split_prospective = _version_split(str(prospective))

            # Shorten the prospective version to be the same length as the spec
            # so that we can determine if the specifier is a prefix of the
            # prospective version or not.
            shortened_prospective = split_prospective[: len(split_spec)]

            # Pad out our two sides with zeros so that they both equal the same
            # length.
            padded_spec, padded_prospective = _pad_version(
                split_spec, shortened_prospective
            )

            return padded_prospective == padded_spec
        else:
            # Convert our spec string into a PythonVersion
            spec_version = PythonVersion(spec)

            # If the specifier does not have a local segment, then we want to
            # act as if the prospective version also does not have a local
            # segment.
            if not spec_version.local:
                prospective = PythonVersion(prospective.public)

            return prospective == spec_version

    def _compare_not_equal(self, prospective: BaseVersion, spec: str) -> bool:
        return not self._compare_equal(prospective, spec)

    def _compare_less_than_equal(self, prospective: BaseVersion,
                                 spec: str) -> bool:
        # NB: Local version identifiers are NOT permitted in the version
        # specifier, so local version labels can be universally removed from
        # the prospective version.
        return PythonVersion(prospective.public) <= PythonVersion(spec)

    def _compare_greater_than_equal(self, prospective: BaseVersion,
                                    spec: str) -> bool:
        # NB: Local version identifiers are NOT permitted in the version
        # specifier, so local version labels can be universally removed from
        # the prospective version.
        return PythonVersion(prospective.public) >= PythonVersion(spec)

    def _compare_less_than(self, prospective: BaseVersion,
                           spec_str: str) -> bool:
        # Convert our spec to a PythonVersion instance, since we'll want to
        # work with it as a version.
        spec = PythonVersion(spec_str)

        # Check to see if the prospective version is less than the spec
        # version. If it's not we can short circuit and just return False now
        # instead of doing extra unneeded work.
        if not prospective < spec:
            return False

        # This special case is here so that, unless the specifier itself
        # includes is a pre-release version, that we do not accept pre-release
        # versions for the version mentioned in the specifier (e.g. <3.1 should
        # not match 3.1.dev0, but should match 3.0.dev0).
        if not spec.is_prerelease and prospective.is_prerelease:
            if ( PythonVersion(prospective.base_version) ==
                 PythonVersion(spec.base_version) ):
                return False

        # If we've gotten to here, it means that prospective version is both
        # less than the spec version *and* it's not a pre-release of the same
        # version in the spec.
        return True

    def _compare_greater_than(self, prospective: BaseVersion,
                              spec_str: str) -> bool:
        # Convert our spec to a PythonVersion instance, since we'll want to
        # work with it as a version.
        spec = PythonVersion(spec_str)

        # Check to see if the prospective version is greater than the spec
        # version. If it's not we can short circuit and just return False now
        # instead of doing extra unneeded work.
        if not prospective > spec:
            return False

        # This special case is here so that, unless the specifier itself
        # includes is a post-release version, that we do not accept
        # post-release versions for the version mentioned in the specifier
        # (e.g. >3.1 should not match 3.0.post0, but should match 3.2.post0).
        if not spec.is_postrelease and prospective.is_postrelease:
            if ( PythonVersion(prospective.base_version) ==
                 PythonVersion(spec.base_version) ):
                return False

        # Ensure that we do not allow a local version of the version mentioned
        # in the specifier, which is technically greater than, to match.
        if prospective.local is not None:
            if ( PythonVersion(prospective.base_version) ==
                 PythonVersion(spec.base_version) ):
                return False

        # If we've gotten to here, it means that prospective version is both
        # greater than the spec version *and* it's not a pre-release of the
        # same version in the spec.
        return True

    def _compare_arbitrary(self, prospective: PythonVersion,
                           spec: str) -> bool:
        return str(prospective).lower() == str(spec).lower()

    @property
    def prereleases(self) -> bool:
        # If there is an explicit prereleases set for this, then we'll just
        # blindly use that.
        if self._prereleases is not None:
            return self._prereleases

        # Look at all of our specifiers and determine if they are inclusive
        # operators, and if they are if they are including an explicit
        # prerelease.
        operator, version = self._spec
        if operator in ["==", ">=", "<=", "~=", "==="]:
            # The == specifier can include a trailing .*, if it does we
            # want to remove before parsing.
            if operator == "==" and version.endswith(".*"):
                version = version[:-2]

            # Parse the version, and if it is a pre-release than this
            # specifier allows pre-releases.
            if PythonVersion(version).is_prerelease:
                return True

        return False

    @prereleases.setter
    def prereleases(self, value: bool) -> None:
        self._prereleases = value


def _canonicalize_version(_version: str) -> str:
    """
    This is very similar to PythonVersion.__str__, but has one subtle
    difference with the way it handles the release segment.
    """

    try:
        version = PythonVersion(_version)
    except InvalidVersion:
        return _version

    parts = []

    # Epoch
    if version.epoch != 0:
        parts.append("{0}!".format(version.epoch))

    # Release segment
    # NB: This strips trailing '.0's to normalize
    parts.append(re.sub(r"(\.0)+$", "", ".".join(
        str(x) for x in version.release
    )))

    # Pre-release
    if version.pre is not None:
        parts.append("".join(str(x) for x in version.pre))

    # Post-release
    if version.post is not None:
        parts.append(".post{0}".format(version.post))

    # Development release
    if version.dev is not None:
        parts.append(".dev{0}".format(version.dev))

    # Local version segment
    if version.local is not None:
        parts.append("+{0}".format(version.local))

    return "".join(parts)


_prefix_regex = re.compile(r"^([0-9]+)((?:a|b|c|rc)[0-9]+)$")


def _version_split(version: str) -> List[str]:
    result: List[str] = []
    for item in version.split("."):
        match = _prefix_regex.search(item)
        if match:
            result.extend(match.groups())
        else:
            result.append(item)
    return result


def _pad_version(
    left: List[str], right: List[str],
) -> Tuple[List[str], List[str]]:
    left_split, right_split = [], []

    # Get the release segment of our versions
    left_split.append(list(itertools.takewhile(lambda x: x.isdigit(), left)))
    right_split.append(list(itertools.takewhile(lambda x: x.isdigit(), right)))

    # Get the rest of our versions
    left_split.append(left[len(left_split[0]):])
    right_split.append(right[len(right_split[0]):])

    # Insert our padding
    left_split.insert(1, ["0"] * max(
        0, len(right_split[0]) - len(left_split[0])
    ))
    right_split.insert(1, ["0"] * max(
        0, len(left_split[0]) - len(right_split[0])
    ))

    return (list(itertools.chain(*left_split)),
            list(itertools.chain(*right_split)))


class PythonSpecifierSet(BaseSpecifierSet):
    def __init__(self, specifiers: str = "",
                 prereleases: Optional[bool] = None) -> None:
        # Split on , to break each individual specifier into its own item, and
        # strip each item to remove leading/trailing whitespace.
        split_specifiers = [s.strip() for s in specifiers.split(",")
                            if s.strip()]

        # Parse each individual specifier.
        parsed: Set[BaseSpecifier] = set()
        for specifier in split_specifiers:
            parsed.add(PythonSpecifier(specifier))

        super().__init__(parsed, prereleases)

    def _coerce_version(self, version: UnparsedVersion) -> PythonVersion:
        if not isinstance(version, PythonVersion):
            version = PythonVersion(str(version))
        return version

    def _filter_prereleases(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool]
    ) -> Iterable[UnparsedVersion]:
        filtered: List[UnparsedVersion] = []
        found_prereleases: List[UnparsedVersion] = []

        for item in iterable:
            # Ensure that we some kind of Version class for this item.
            parsed_version = self._coerce_version(item)

            # Store any item which is a pre-release for later unless we've
            # already found a final version or we are accepting prereleases
            if parsed_version.is_prerelease and not prereleases:
                if not filtered:
                    found_prereleases.append(item)
            else:
                filtered.append(item)

        # If we've found no items except for pre-releases, then we'll go
        # ahead and use the pre-releases
        if not filtered and found_prereleases and prereleases is None:
            return found_prereleases

        return filtered


Version = PythonVersion
Specifier = PythonSpecifier
SpecifierSet = PythonSpecifierSet
