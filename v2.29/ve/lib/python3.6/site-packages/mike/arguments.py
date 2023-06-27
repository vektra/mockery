from argparse import *

_ArgumentParser = ArgumentParser
_Action = Action


# Add some simple wrappers to make it easier to specify shell-completion
# behaviors.

def _add_complete(argument, complete):
    if complete is not None:
        argument.complete = complete
    return argument


class Action(_Action):
    def __init__(self, *args, complete=None, **kwargs):
        super().__init__(*args, **kwargs)
        _add_complete(self, complete)


class ArgumentParser(_ArgumentParser):
    @staticmethod
    def _wrap_complete(action):
        def wrapper(*args, complete=None, **kwargs):
            return _add_complete(action(*args, **kwargs), complete)

        return wrapper

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        for k, v in self._registries['action'].items():
            self._registries['action'][k] = self._wrap_complete(v)
