// This package contains implementation for checking if an item has been seen previously. This is different from a bloom-filter
// as it prioritizes false-negatives (return absence even if the object was seen previously), instead of false-positives.
// A Bloom filter prioritizes false-positives (return presence even if the object was not seen previously)
// and thus is able to provide a high compression.
// This is similar to a cache, with the difference being that it does not store a value and can use other optimized
// representations to compress the keys and optimize for storage.
// The simplest implementation for the fastcheck.Filter is an LRU cache
// the package provides alternative implementations and is inspired by the code in
// https://github.com/jmhodges/opposite_of_a_bloom_filter/ and
// the related blog https://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/
package fastcheck
