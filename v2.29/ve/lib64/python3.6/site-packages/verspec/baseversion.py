import abc

from typing import Any, Optional, Union, Tuple

CmpKey = Tuple[Any, ...]
LetterVersion = Tuple[str, int]


class InvalidVersion(ValueError):
    """
    An invalid version was found, users should refer to PEP 440.
    """


class BaseVersion(metaclass=abc.ABCMeta):
    _key = None  # type: CmpKey

    def __hash__(self) -> int:
        return hash(self._key)

    def __repr__(self) -> str:
        return "<{}({})>".format(type(self).__name__, repr(str(self)))

    def __lt__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key < other._key

    def __le__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key <= other._key

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key == other._key

    def __ge__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key >= other._key

    def __gt__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key > other._key

    def __ne__(self, other: object) -> bool:
        if not isinstance(other, type(self)):
            return NotImplemented

        return self._key != other._key

    @property
    @abc.abstractmethod
    def public(self) -> str:
        pass

    @property
    @abc.abstractmethod
    def base_version(self) -> str:
        pass

    @property
    @abc.abstractmethod
    def epoch(self) -> int:
        pass

    @property
    @abc.abstractmethod
    def release(self) -> Optional[Tuple[int, ...]]:
        pass

    @property
    @abc.abstractmethod
    def pre(self) -> Optional[LetterVersion]:
        pass

    @property
    @abc.abstractmethod
    def post(self) -> Optional[int]:
        pass

    @property
    @abc.abstractmethod
    def dev(self) -> Optional[int]:
        pass

    @property
    @abc.abstractmethod
    def local(self) -> Optional[str]:
        pass

    @property
    @abc.abstractmethod
    def is_prerelease(self) -> bool:
        pass

    @property
    @abc.abstractmethod
    def is_postrelease(self) -> bool:
        pass

    @property
    @abc.abstractmethod
    def is_devrelease(self) -> bool:
        pass


UnparsedVersion = Union[BaseVersion, str]
