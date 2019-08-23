import { colors } from '../theme';

export default [
    {
        fill: colors.gray1,
        id: 'start-node'
    },
    {
        fill: colors.gold,
        id: 'batch-test',
        parentIds: ['start-node']
    },
    {
        fill: colors.indigo,
        id: 'print-always',
        parentIds: ['print-sum', 'start-node']
    },
    {
        fill: colors.turqoise,
        id: 'print-spark',
        parentIds: ['sparky']
    },
    {
        fill: colors.sepia,
        id: 'print-sum',
        parentIds: ['print2', 'print4', 'print3']
    },
    {
        fill: colors.sepia,
        id: 'print-sum2',
        parentIds: ['batch-test']
    },
    {
        fill: colors.indigo,
        id: 'print1a',
        parentIds: ['start-node']
    },
    {
        fill: colors.indigo,
        id: 'print1b',
        parentIds: ['start-node']
    },
    {
        fill: colors.indigo,
        id: 'print2',
        parentIds: ['print1a', 'print1b']
    },
    {
        fill: colors.indigo,
        id: 'print3',
        parentIds: ['print2']
    },
    {
        fill: colors.cobalt,
        id: 'print4',
        parentIds: ['print3']
    },
    {
        fill: colors.turqoise,
        id: 'sparky',
        parentIds: ['print2', 'print4', 'print3']
    }
];
