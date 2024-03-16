from tests.flytekit.unit.core.tracker.a import A

b_local_a = A()


def get_a():
    return A()


class B(A):
    ...


local_b = B()
