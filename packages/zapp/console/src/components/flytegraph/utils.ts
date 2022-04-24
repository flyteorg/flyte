import { Point } from './types';

/** Simple helper to get the midpoint between two points a & b */
export function getMidpoint(a: Point, b: Point) {
  return {
    x: (a.x + b.x) / 2,
    y: (a.y + b.y) / 2,
  };
}

interface PointBucket<T extends Point> {
  x: number;
  points: T[];
}

/** Will organize a set of points into buckets based on a common
 * x/y value.
 */
export function groupBy<T extends Point>(prop: 'x' | 'y', points: T[]): PointBucket<T>[] {
  points.sort((a, b) => a[prop] - b[prop]);
  return points.reduce<PointBucket<T>[]>((buckets, point, idx) => {
    const previousPoint = idx === 0 ? null : points[idx - 1];
    if (previousPoint === null || Math.abs(point[prop] - previousPoint[prop]) > Number.EPSILON) {
      const newBucket = {
        x: point[prop],
        points: [point],
      };
      buckets.push(newBucket);
    } else {
      const bucket = buckets[buckets.length - 1];
      bucket.points.push(point);
    }
    return buckets;
  }, []);
}
