export enum ExecutionMetadataLabels {
    cluster = 'Cluster',
    domain = 'Domain',
    duration = 'Duration',
    time = 'Time',
    relatedTo = 'Related to',
    version = 'Version',
    serviceAccount = 'Service Account',
    rawOutputPrefix = 'Raw Output Prefix',
    parallelism = 'Parallelism'
}

export const tabs = {
    nodes: {
        id: 'nodes',
        label: 'Nodes'
    },
    graph: {
        id: 'graph',
        label: 'Graph'
    }
};

export const executionActionStrings = {
    clone: 'Clone Execution'
};

export const backLinkTitle = 'Back to parent';
