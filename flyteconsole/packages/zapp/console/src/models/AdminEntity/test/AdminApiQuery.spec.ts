import { generateAdminApiQuery } from '../AdminApiQuery';
import { sortQueryKeys } from '../constants';
import { FilterOperation, FilterOperationName, Sort, SortDirection } from '../types';

function makeFilter(key: string, operation: FilterOperationName, value: string): FilterOperation {
  return { key, operation, value };
}

describe('generateAdminApiQuery', () => {
  it('passes through token and limit values', () => {
    const query = { token: 'nextItem', limit: 5 };
    expect(generateAdminApiQuery(query)).toEqual(query);
  });

  describe('filters', () => {
    it('handles CONTAINS filter', () => {
      const query = {
        filter: [makeFilter('searchMe', FilterOperationName.CONTAINS, 'search')],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'contains(searchMe,search)',
      });
    });

    it('handles EQ filter', () => {
      const query = {
        filter: [makeFilter('it', FilterOperationName.EQ, 'matches')],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'eq(it,matches)',
      });
    });

    it('handles GTE filter', () => {
      const query = {
        filter: [makeFilter('aNumber', FilterOperationName.GTE, '1')],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'gte(aNumber,1)',
      });
    });

    it('handles LTE filter', () => {
      const query = {
        filter: [makeFilter('aNumber', FilterOperationName.LTE, '1')],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'lte(aNumber,1)',
      });
    });

    it('handles NE filter', () => {
      const query = {
        filter: [makeFilter('theSame', FilterOperationName.NE, 'notTheSame')],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'ne(theSame,notTheSame)',
      });
    });

    it('concatenates multiple filters', () => {
      const query = {
        filter: [
          makeFilter('rangeValue', FilterOperationName.GTE, '1'),
          makeFilter('rangeValue', FilterOperationName.LTE, '5'),
        ],
      };
      expect(generateAdminApiQuery(query)).toEqual({
        filters: 'gte(rangeValue,1)+lte(rangeValue,5)',
      });
    });

    it('throws if passed an unsupported filter operation', () => {
      const query = {
        filter: [makeFilter('key', 'badOp' as FilterOperationName, 'value')],
      };
      expect(() => generateAdminApiQuery(query)).toThrow();
    });
  });

  describe('sort', () => {
    it('generates correct ascending filter', () => {
      const sort: Sort = {
        direction: SortDirection.ASCENDING,
        key: 'sortKey',
      };
      expect(generateAdminApiQuery({ sort })).toEqual({
        [sortQueryKeys.direction]: 'ASCENDING',
        [sortQueryKeys.key]: 'sortKey',
      });
    });

    it('generates correct descending filter', () => {
      const sort: Sort = {
        direction: SortDirection.DESCENDING,
        key: 'sortKey',
      };
      expect(generateAdminApiQuery({ sort })).toEqual({
        [sortQueryKeys.direction]: 'DESCENDING',
        [sortQueryKeys.key]: 'sortKey',
      });
    });
  });
});
