import abc
from typing import (Callable, Dict, Iterable, Iterator, Optional, Pattern, Set,
                    Tuple, Union)

from .baseversion import BaseVersion, UnparsedVersion


CallableOperator = Callable[[BaseVersion, str], bool]


class InvalidSpecifier(ValueError):
    """
    An invalid specifier was found, users should refer to PEP 440.
    """


class BaseSpecifier(metaclass=abc.ABCMeta):
    @abc.abstractmethod
    def __str__(self) -> str:
        """
        Returns the str representation of this Specifier like object. This
        should be representative of the Specifier itself.
        """

    @abc.abstractmethod
    def __hash__(self) -> int:
        """
        Returns a hash value for this Specifier like object.
        """

    @abc.abstractmethod
    def __eq__(self, other: object) -> bool:
        """
        Returns a boolean representing whether or not the two Specifier like
        objects are equal.
        """

    @abc.abstractmethod
    def __ne__(self, other: object) -> bool:
        """
        Returns a boolean representing whether or not the two Specifier like
        objects are not equal.
        """

    @abc.abstractproperty
    def prereleases(self) -> Optional[bool]:
        """
        Returns whether or not pre-releases as a whole are allowed by this
        specifier.
        """

    @prereleases.setter
    def prereleases(self, value: bool) -> None:
        """
        Sets whether or not pre-releases as a whole are allowed by this
        specifier.
        """

    @abc.abstractmethod
    def contains(self, item: UnparsedVersion,
                 prereleases: Optional[bool] = None) -> bool:
        """
        Determines if the given item is contained within this specifier.
        """

    @abc.abstractmethod
    def filter(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool] = None,
    ) -> Iterable[UnparsedVersion]:
        """
        Takes an iterable of items and filters them so that only items which
        are contained within this specifier are allowed in it.
        """


class IndividualSpecifier(BaseSpecifier, metaclass=abc.ABCMeta):
    _operators: Dict[str, str] = {}
    _regex = None  # type: Pattern

    def __init__(self, spec: str = "",
                 prereleases: Optional[bool] = None) -> None:
        match = self._regex.search(spec)
        if not match:
            raise InvalidSpecifier("Invalid specifier: '{0}'".format(spec))

        self._spec: Tuple[str, str] = (
            match.group("operator").strip(),
            match.group("version").strip(),
        )

        # Store whether or not this Specifier should accept prereleases
        self._prereleases = prereleases

    @abc.abstractmethod
    def _coerce_version(self, version: UnparsedVersion) -> BaseVersion:
        pass

    @property
    @abc.abstractmethod
    def _canonical_spec(self) -> Tuple[str, UnparsedVersion]:
        pass

    def __repr__(self) -> str:
        pre = (
            ", prereleases={0!r}".format(self.prereleases)
            if self._prereleases is not None
            else ""
        )

        return "<{0}({1!r}{2})>".format(type(self).__name__, str(self), pre)

    def __str__(self) -> str:
        return "{0}{1}".format(*self._spec)

    def __hash__(self) -> int:
        return hash(self._canonical_spec)

    def __eq__(self, other: object) -> bool:
        if isinstance(other, str):
            try:
                other = type(self)(str(other))
            except InvalidSpecifier:
                return NotImplemented
        elif not isinstance(other, type(self)):
            return NotImplemented

        return self._canonical_spec == other._canonical_spec

    def __ne__(self, other: object) -> bool:
        if isinstance(other, str):
            try:
                other = type(self)(str(other))
            except InvalidSpecifier:
                return NotImplemented
        elif not isinstance(other, type(self)):
            return NotImplemented

        return self._canonical_spec != other._canonical_spec

    def _get_operator(self, op: str) -> CallableOperator:
        operator_callable: CallableOperator = getattr(
            self, "_compare_{0}".format(self._operators[op])
        )
        return operator_callable

    @property
    def operator(self) -> str:
        return self._spec[0]

    @property
    def version(self) -> str:
        return self._spec[1]

    @property
    def prereleases(self) -> Optional[bool]:
        return self._prereleases

    @prereleases.setter
    def prereleases(self, value: bool) -> None:
        self._prereleases = value

    def __contains__(self, item: UnparsedVersion) -> bool:
        return self.contains(item)

    def contains(self, item: UnparsedVersion,
                 prereleases: Optional[bool] = None) -> bool:

        # Determine if prereleases are to be allowed or not.
        if prereleases is None:
            prereleases = self.prereleases

        # Normalize item to a Version or LooseVersion, this allows us to have
        # a shortcut for ``"2.0" in Specifier(">=2")
        normalized_item = self._coerce_version(item)

        # Determine if we should be supporting prereleases in this specifier
        # or not, if we do not support prereleases than we can short circuit
        # logic if this version is a prereleases.
        if normalized_item.is_prerelease and not prereleases:
            return False

        # Actually do the comparison to determine if this item is contained
        # within this Specifier or not.
        operator_callable = self._get_operator(self.operator)
        return operator_callable(normalized_item, self.version)

    def filter(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool] = None,
    ) -> Iterable[UnparsedVersion]:
        yielded = False
        found_prereleases = []

        kw = {"prereleases": prereleases if prereleases is not None else True}

        # Attempt to iterate over all the values in the iterable and if any of
        # them match, yield them.
        for version in iterable:
            parsed_version = self._coerce_version(version)

            if self.contains(parsed_version, **kw):
                # If our version is a prerelease, and we were not set to allow
                # prereleases, then we'll store it for later incase nothing
                # else matches this specifier.
                if parsed_version.is_prerelease and not (
                    prereleases or self.prereleases
                ):
                    found_prereleases.append(version)
                # Either this is not a prerelease, or we should have been
                # accepting prereleases from the beginning.
                else:
                    yielded = True
                    yield version

        # Now that we've iterated over everything, determine if we've yielded
        # any values, and if we have not and we have any prereleases stored up
        # then we will go ahead and yield the prereleases.
        if not yielded and found_prereleases:
            for version in found_prereleases:
                yield version


class BaseSpecifierSet(BaseSpecifier, metaclass=abc.ABCMeta):
    def __init__(self, parsed_specifiers: Set[BaseSpecifier],
                 prereleases: Optional[bool]) -> None:
        # Turn our parsed specifiers into a frozen set and save them for later.
        self._specs = frozenset(parsed_specifiers)

        # Store our prereleases value so we can use it later to determine if
        # we accept prereleases or not.
        self._prereleases = prereleases

    @abc.abstractmethod
    def _coerce_version(self, version: UnparsedVersion) -> BaseVersion:
        pass

    @abc.abstractmethod
    def _filter_prereleases(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool]
    ) -> Iterable[UnparsedVersion]:
        pass

    def __repr__(self) -> str:
        pre = (
            ", prereleases={0!r}".format(self.prereleases)
            if self._prereleases is not None
            else ""
        )

        return "<{0}({1!r}{2})>".format(type(self).__name__, str(self), pre)

    def __str__(self) -> str:
        return ",".join(sorted(str(s) for s in self._specs))

    def __hash__(self) -> int:
        return hash(self._specs)

    def __and__(
        self, other: Union['BaseSpecifierSet', str],
    ) -> 'BaseSpecifierSet':
        if isinstance(other, str):
            other = type(self)(other)  # type: ignore
        elif not isinstance(other, type(self)):
            # Currently, SpecifierSets and LooseSpecifierSets can't be
            # combined.
            return NotImplemented

        specifier = type(self)()  # type: ignore
        specifier._specs = frozenset(self._specs | other._specs)

        if self._prereleases is None and other._prereleases is not None:
            specifier._prereleases = other._prereleases
        elif self._prereleases is not None and other._prereleases is None:
            specifier._prereleases = self._prereleases
        elif self._prereleases == other._prereleases:
            specifier._prereleases = self._prereleases
        else:
            raise ValueError(
                "Cannot combine {}s with True and False prerelease "
                "overrides.".format(type(self).__name__)
            )

        return specifier

    def __eq__(self, other: object) -> bool:
        if isinstance(other, (str, IndividualSpecifier)):
            other = type(self)(str(other))  # type: ignore
        elif not isinstance(other, BaseSpecifierSet):
            return NotImplemented

        return self._specs == other._specs

    def __ne__(self, other: object) -> bool:
        if isinstance(other, (str, IndividualSpecifier)):
            other = type(self)(str(other))  # type: ignore
        elif not isinstance(other, BaseSpecifierSet):
            return NotImplemented

        return self._specs != other._specs

    def __len__(self) -> int:
        return len(self._specs)

    def __iter__(self) -> Iterator[BaseSpecifier]:
        return iter(self._specs)

    @property
    def prereleases(self) -> Optional[bool]:
        # If we have been given an explicit prerelease modifier, then we'll
        # pass that through here.
        if self._prereleases is not None:
            return self._prereleases

        # If we don't have any specifiers, and we don't have a forced value,
        # then we'll just return None since we don't know if this should have
        # pre-releases or not.
        if not self._specs:
            return None

        # Otherwise we'll see if any of the given specifiers accept
        # prereleases, if any of them do we'll return True, otherwise False.
        return any(s.prereleases for s in self._specs)

    @prereleases.setter
    def prereleases(self, value: bool) -> None:
        self._prereleases = value

    def __contains__(self, item: UnparsedVersion) -> bool:
        return self.contains(item)

    def contains(self, item: UnparsedVersion,
                 prereleases: Optional[bool] = None) -> bool:
        # Ensure that our item is a PythonVersion or LooseVersion instance.
        parsed_item = self._coerce_version(item)

        # Determine if we're forcing a prerelease or not, if we're not forcing
        # one for this particular filter call, then we'll use whatever the
        # SpecifierSet thinks for whether or not we should support prereleases.
        if prereleases is None:
            prereleases = self.prereleases

        # We can determine if we're going to allow pre-releases by looking to
        # see if any of the underlying items supports them. If none of them do
        # and this item is a pre-release then we do not allow it and we can
        # short circuit that here.
        # Note: This means that 1.0.dev1 would not be contained in something
        #       like >=1.0.devabc however it would be in >=1.0.debabc,>0.0.dev0
        if not prereleases and parsed_item.is_prerelease:
            return False

        # We simply dispatch to the underlying specs here to make sure that the
        # given version is contained within all of them.
        # Note: This use of all() here means that an empty set of specifiers
        #       will always return True, this is an explicit design decision.
        return all(
            s.contains(parsed_item, prereleases=prereleases)
            for s in self._specs
        )

    def filter(
        self, iterable: Iterable[UnparsedVersion],
        prereleases: Optional[bool] = None,
    ) -> Iterable[UnparsedVersion]:
        # Determine if we're forcing a prerelease or not, if we're not forcing
        # one for this particular filter call, then we'll use whatever the
        # SpecifierSet thinks for whether or not we should support prereleases.
        if prereleases is None:
            prereleases = self.prereleases

        # If we have any specifiers, then we want to wrap our iterable in the
        # filter method for each one, this will act as a logical AND amongst
        # each specifier.
        if self._specs:
            for spec in self._specs:
                iterable = spec.filter(iterable, prereleases=bool(prereleases))
            return iterable
        # If we do not have any specifiers, then we need to have a rough filter
        # which will filter out any pre-releases, unless there are no final
        # releases, and which will filter out LooseVersion in general.
        else:
            return self._filter_prereleases(iterable, prereleases)
