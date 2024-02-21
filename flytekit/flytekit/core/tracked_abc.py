from abc import ABC

from flytekit.core.tracker import TrackedInstance


class FlyteTrackedABC(type(TrackedInstance), type(ABC)):  # type: ignore
    """
    This class exists because if you try to inherit from abc.ABC and TrackedInstance by itself, you'll get the
    well-known ``TypeError: metaclass conflict: the metaclass of a derived class must be a (non-strict) subclass
    of the metaclasses of all its bases`` error.
    """
