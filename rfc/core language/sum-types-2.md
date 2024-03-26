# Sum Types (Unions) - Follow Up

**This is a follow-up to:** [the Sum Types RFC](https://github.com/maximsmol/flyte/blob/master/rfc/core%20language/sum-types.md)

# Executive Summary

Some questions on the previously proposed implementations made it clear that a deeper investigation into possible alternatives was required. I consider 3 programming languages with fundamentally different union type implementations in the core language or the standard library. I develop a new version of the sum type IDL representation that accommodates all three languages.

# Examples in Programming Languages

## Python

The type-erased case.

- No runtime representation
- The union is a set of types
  - Duplicates get collapsed
  - Single-type unions collapse to the underlying type
- Order of types does not matter
  - Cannot effect runtime behavior as there is no tag
  - Does not effect type equality

```py
>>> # No runtime representation
>>> a : t.Union[str, int] = 10
>>> a
10

>>> # Single-type union collapse
>>> t.Union[int, int]
<class 'int'>

>>> # Trivial duplicate collapse
>>> t.Union[int, int, str]
typing.Union[int, str]

>>> # Non-trivial duplicate collapse:
>>> a = t.Union[t.List[t.Union[str, int]], t.List[t.Union[str, int]]]
>>> b = t.Union[t.List[t.Union[str, int]], t.List[t.Union[int, str]]]
>>> a
typing.List[typing.Union[str, int]]
>>> b
typing.List[typing.Union[str, int]]
>>> a == b
True

>>> # Order does not matter:
>>> t.Union[str, int] == t.Union[int, str]
True
```

## Haskell

Algebraic types case.

- Runtime representation carries symbolic tag
- The union is a set of (tag, type) tuples
  - Duplicate tags are a compile-time error
  - Duplicate types with different tags are allowed
- Order of variants does not matter as the symbolic tag does not depend on its order in the union
  - Defining the same union type but with different order of variants in different compilation unions (when compiling separately from linking) works as expected as the symbols referring to the type contain the symbolic tags

```haskell
data Test = Hello Int | World String

-- Runtime value carries tag, can be introspected
case x of
  Hello a -> "Found int: " ++ show a
  World b -> "Found string: " ++ show b

-- Duplicates are a compile-time error
data Test1 = Hello | Hello
{-|
test.hs:1:21: error:
  Multiple declarations of ‘Hello’
  Declared at: test.hs:1:13
               test.hs:1:21
-}

-- Duplicate types are allowed
data Test2 = Left Int | Right Int
case x of
  Left a -> "Found int: " ++ show a
  Right b -> "Found a different int: " ++ show b
```

## C++ (std::variant)

In-between case (indexed union case).

- Runtime representation carries a positional tag
- Two available APIs
  - One uses positional indexes as the tag
  - One uses the types themselves as tags
- The union is a list of types
  - Duplicate types cannot be used with the type-indexed API (unless unambiguous) but can be used with the position-indexed API
- Order of variants matters
  - Can influence runtime behavior
  - Distinct order means distinct types

```cpp
// Runtime representation carries a position tag, showing both APIs
std::variant<int, bool> a = 10;
assert(std::get<int>(a) == 10);
assert(std::get<0>(a) == 10);
a = false;
assert(std::get<bool>(a) == false);
assert(std::get<1>(a) == false);

// Failure cases
std::get<2>(a); // no matching function for call to 'get'
std::get<double>(a); // no matching function for call to 'get'

std::get<0>(a);
/*
terminate called after throwing an instance of 'std::bad_variant_access'
  what():  Unexpected index
*/

// Duplicate types are allowed but must use an unambiguous API
std::variant<int, int> b = 10;
/*
no viable conversion from 'int' to
    'std::variant<int, int>'
*/

// Unambiguous uses allow both APIs
std::variant<int, int, bool> c = false;

assert(std::get<bool>(c) == false);
assert(std::get<2>(c) == false);

// Ambiguous uses of the API do not work, the index-based API is the never ambiguous
c.emplace<0>(10); // Assignment using the index-based API
std::get<int>(c);
/*
error: 
      static_assert failed due to requirement
      '__detail::__variant::__exactly_once<int, int, int, bool>' "T should occur
      for exactly once in alternatives"
      static_assert(__detail::__variant::__exactly_once<_Tp, _Types...>,
*/

std::get<0>(c) == 10;

// Order of types matters
if (c.index() == 0)
  std::cout << "First integer" << std::endl;
else if (c.index() == 1)
  std::cout << "Second integer" << std::endl;
else if (c.index() == 2)
  std::cout << "Boolean" << std::endl;

std::variant<int, bool> x = false;
std::variant<bool, int> y = true;

x = y;
/*
error: no viable overloaded '='
  x = y;
  ~ ^ ~
*/
```

# Design considerations

## Tagged vs Untagged

- To properly support languages like Haskell and C++ the backend representation of union types should use tags of some kind.

## Type of Tag

- First-class Haskell support needs string tags
  - Integer tags open the possibility of incorrectly (according to language semantics) decoding a union-typed literal from IDL to a Haskell type if the Haskell source code changes to rearrange the variants
  - Such a change is a no-op according to language semantics so this is an issue unless it is guaranteed that IDL `LiteralRepr`s are never serialized (or are never reused across task/workflow versions)
  - This is also an issue when linking Haskell object files since they remain compatible even when produced from these two different (but equivalent) versions of the source code
  - Both of these error cases are unlikely to occur naturally but not supporting them would cause obscure issues for users
- First-class C++ support needs integer tags

Since integers can be stringified, string tags offer first-class support for both languages

## Tags in Python

Since Python's `typing.Union` is untagged, it could be implemented without a tag and even without a `LiteralRepr` for the union values. Here is why this is not a good choice:

- An untagged implementation requires being able to determine whether a given `LiteralRepr` is castable to a given Python type. The current type transformer implementation is not designed as a type validator and may not throw on incompatible types during the transformation at all, or will throw an error indistinguishable from a normal Python programmer-error exception
- Another issue is that in case a custom class was defined (e.g. `MyInt` which is simply a proxy for the default `int`), then the `LiteralRepr` for an integer value would convert into both `MyInt` and `int` without an error so the choice between the two would be ambiguous and the runtime behavior could be influenced since the custom type transformer for `MyInt` can run arbitrary code

We could use an index-based tag, but there are multiple reasons why that is not a good idea either:

- `typing.Union` semantics are that of a set (the order does not matter for equality comparison, duplicates are eliminated), though in practice it the source ordering of the variants can be recovered in cases without duplicates. Note that the CPython implementation actually uses a tuple behind the scenes, so this is more about intent and less about factual behavior of the class. The [PEP](https://www.python.org/dev/peps/pep-0484/#union-types) also specifies that Union is expected to accept a "set" of types, though it is unclear whether this is referring to a specific data structure or just a figure of speech
- If we ignore the apparent intent to implement unions as sets, the issue of code refactoring arises again as changing the order of variants in a union should not effect behavior. This goes away if we can guarantee that IDL is never serialized or that the serialized messages are never reused across different task/workflow versions

These problems compound with the requirements necessary for properly supporting Haskell and C++.

# Proposed Implementation

Use a string tag. In Haskell use the symbolic tag. In C++ use the index (serialized to a string) as the tag. In Python use the name of the type transformer (already present on all transformers).

In Python the correspondence between a tag and the choice of type transformer must be 1-to-1 i.e. type transformer names must be made unique which is already the case but not formalized.

The matching procedure for Haskell and C++ is trivial. In Python, we must deal specially with duplicates. For example:

```py
from typing import Union

# MyList is a proxy type similar to MyInt - a no-op wrapper around the native
# Python type but with a custom type transformer (which might have side effects)
def f(x: Union[MyList[Union[MyInt, int]], MyList[int]]):
  ...
```

In this case, having the `MyList` tag is not enough to disambiguate the choice of variant. We iterate each candidate variant and try to match it recursively with the literal. It is not required that all type transformers can fail gracefully when given a value of any incorrect type, only that they fail on union `LiteralRepr`s and that the union type transformer can recognize its own literals. The only possible difference between how the types are resolved is in the choice of type transformers, which is made completely unambiguous by traversing the type tree with its tags (since the tags resolve the only ambiguity). This procedure thus guarantees a value will be recovered appropriately from a union IDL representation.
