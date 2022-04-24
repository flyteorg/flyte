declare module 'd3-dag' {
  export interface DierarchyLink<Datum> {
    /**
     * The source of the link.
     */
    source: DierarchyNode<Datum>;

    /**
     * The target of the link.
     */
    target: DierarchyNode<Datum>;
  }

  export interface DierarchyNode<Datum> {
    /**
     * The associated data, as specified to the constructor.
     */
    data: Datum;

    /**
     * Zero for the root node, and increasing by one for each descendant generation.
     */
    readonly depth: number;

    /**
     * Zero for leaf nodes, and the greatest distance from any descendant leaf for internal nodes.
     */
    readonly height: number;

    /**
     * The parent node, or null for the root node.
     */
    parent: this | null;

    /**
     * An array of child nodes, if any; undefined for leaf nodes.
     */
    children?: this[];

    /**
     * Aggregated numeric value as calculated by `sum(value)` or `count()`, if previously invoked.
     */
    readonly value?: number;

    /**
     * Optional node id string set by `StratifyOperator`, if hierarchical data was created from tabular data using stratify().
     */
    readonly id?: string;

    /**
     * Returns the array of ancestors nodes, starting with this node, then followed by each parent up to the root.
     */
    ancestors(): this[];

    /**
     * Returns the array of descendant nodes, starting with this node, then followed by each child in topological order.
     */
    descendants(): this[];

    /**
     * Returns the array of leaf nodes in traversal order; leaves are nodes with no children.
     */
    leaves(): this[];

    /**
     * Returns the shortest path through the hierarchy from this node to the specified target node.
     * The path starts at this node, ascends to the least common ancestor of this node and the target node, and then descends to the target node.
     *
     * @param target The target node.
     */
    path(target: this): this[];

    /**
     * Returns an array of links for this node, where each link is an object that defines source and target properties.
     * The source of each link is the parent node, and the target is a child node.
     */
    links(): DierarchyLink<Datum>[];

    /**
     * Evaluates the specified value function for this node and each descendant in post-order traversal, and returns this node.
     * The `node.value` property of each node is set to the numeric value returned by the specified function plus the combined value of all descendants.
     *
     * @param value The value function is passed the node’s data, and must return a non-negative number.
     */
    sum(value: (d: Datum) => number): this;

    /**
     * Computes the number of leaves under this node and assigns it to `node.value`, and similarly for every descendant of node.
     * If this node is a leaf, its count is one. Returns this node.
     */
    count(): this;

    /**
     * Sorts the children of this node, if any, and each of this node’s descendants’ children,
     * in pre-order traversal using the specified compare function, and returns this node.
     *
     * @param compare The compare function is passed two nodes a and b to compare.
     * If a should be before b, the function must return a value less than zero;
     * if b should be before a, the function must return a value greater than zero;
     * otherwise, the relative order of a and b are not specified. See `array.sort` for more.
     */
    sort(compare: (a: this, b: this) => number): this;

    /**
     * Invokes the specified function for node and each descendant in breadth-first order,
     * such that a given node is only visited if all nodes of lesser depth have already been visited,
     * as well as all preceding nodes of the same depth.
     *
     * @param func The specified function is passed the current node.
     */
    each(func: (node: this) => void): this;

    /**
     * Invokes the specified function for node and each descendant in post-order traversal,
     * such that a given node is only visited after all of its descendants have already been visited.
     *
     * @param func The specified function is passed the current node.
     */
    eachAfter(func: (node: this) => void): this;

    /**
     * Invokes the specified function for node and each descendant in pre-order traversal,
     * such that a given node is only visited after all of its ancestors have already been visited.
     *
     * @param func The specified function is passed the current node.
     */
    eachBefore(func: (node: this) => void): this;

    /**
     * Return a deep copy of the subtree starting at this node. The returned deep copy shares the same data, however.
     * The returned node is the root of a new tree; the returned node’s parent is always null and its depth is always zero.
     */
    copy(): this;
  }

  export interface DierarchyPointLinkData {
    points: { x: number; y: number }[];
  }

  export interface DierarchyPointLink<Datum> {
    data: DierarchyPointLinkData;
    /**
     * The source of the link.
     */
    source: DierarchyPointNode<Datum>;

    /**
     * The target of the link.
     */
    target: DierarchyPointNode<Datum>;
  }

  export interface DierarchyPointNode<Datum> extends DierarchyNode<Datum> {
    /**
     * The x-coordinate of the node.
     */
    x: number;

    /**
     * The y-coordinate of the node.
     */
    y: number;

    /**
     * Returns an array of links for this node, where each link is an object that defines source and target properties.
     * The source of each link is the parent node, and the target is a child node.
     */
    links(): DierarchyPointLink<Datum>[];
  }

  /**
   * Constructs a root node from the specified hierarchical data.
   *
   * @param data The root specified data.
   * @param children The specified children accessor function invoked for each datum, starting with the root data.
   * Must return an array of data representing the children, and return null or undefined if the current datum has no children.
   * If children is not specified, it defaults to: `(d) => d.children`.
   */
  export function dagHierarchy<Datum>(
    data: Datum,
    children?: (d: Datum) => Datum[] | null | undefined,
  ): DierarchyNode<Datum>;

  // -----------------------------------------------------------------------
  // Dratify
  // -----------------------------------------------------------------------

  export interface DratifyOperator<Datum> {
    /**
     * Generates a new hierarchy from the specified tabular data. Each node in the returned object has a shallow copy of the properties
     * from the corresponding data object, excluding the following reserved properties: id, parentId, children.
     *
     * @param data The root specified data.
     * @throws Error on missing id, ambiguous id, cycle, multiple roots or no root.
     */
    (data: Datum[]): DierarchyNode<Datum>;

    /**
     * Returns the current id accessor, which defaults to: `(d) => d.id`.
     */
    id(): (d: Datum, i: number, data: Datum[]) => string | null | '' | undefined;
    /**
     * Sets the id accessor to the given function.
     * The id accessor is invoked for each element in the input data passed to the dratify operator.
     * The returned string is then used to identify the node's relationships in conjunction with the parent id.
     * For leaf nodes, the id may be undefined, null or the empty string; otherwise, the id must be unique.
     *
     * @param id The id accessor.
     */
    id(id: (d: Datum, i: number, data: Datum[]) => string | null | '' | undefined): this;

    /**
     * Returns the current parent id accessor, which defaults to: `(d) => d.parentId`.
     */
    parentId(): (d: Datum, i: number, data: Datum[]) => string | null | '' | undefined;
    /**
     * Sets the parent id accessor to the given function.
     * The parent id accessor is invoked for each element in the input data passed to the dratify operator.
     * The returned string is then used to identify the node's relationships in conjunction with the id.
     * For the root node, the parent id should be undefined, null or the empty string.
     * There must be exactly one root node in the input data, and no circular relationships.
     *
     * @param parentId The parent id accessor.
     */
    parentId(
      parentId: (d: Datum, i: number, data: Datum[]) => string | null | '' | undefined,
    ): this;
  }

  /**
   * Constructs a new dratify operator with the default settings.
   */
  export function dagStratify<Datum>(): DratifyOperator<Datum>;

  export interface Decross<Datum> {
    (root: DierarchyNode<Datum>): DierarchyPointNode<Datum>;
  }

  export type LayeringAccessor<Datum> = (dag: any) => any;
  export type DecrossAccessor<Datum> = (layers: any) => any;
  export type CoordAccessor<Datum> = (layers: any) => any;

  export function decrossOpt<Datum>();
  export function decrossTwoLayer<Datum>();
  export function layeringCoffmanGraham<Datum>();
  export function coordGreedy<Datum>();
  export function coordVert<Datum>();
  export function coordMinCurve<Datum>();
  export function twolayerOpt<Datum>();

  export interface SugiyamaLayout<Datum> {
    (root: DierarchyNode<Datum>): DierarchyPointNode<Datum>;

    size(): [number, number] | null;
    size(size: [number, number]): this;

    layering(): LayeringAccessor<Datum>;
    layering(layering: LayeringAccessor<Datum>): this;

    decross(): DecrossAccessor<Datum>;
    decross(decross: DecrossAccessor<Datum>): this;

    coord(): CoordAccessor<Datum>;
    coord(coord: CoordAccessor<Datum>): this;

    separation(): (a: DierarchyPointNode<Datum>, b: DierarchyPointNode<Datum>) => number;
    separation(
      separation: (a: DierarchyPointNode<Datum>, b: DierarchyPointNode<Datum>) => number,
    ): this;
  }

  /**
   * Creates a new Sugiyama layout with default settings.
   */
  export function sugiyama<Datum>(): SugiyamaLayout<Datum>;
}
