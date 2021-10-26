import { useState } from 'react';

export function useChartState() {
    const [chartIds, setChartIds] = useState<string[]>([]);
    const onToggle = (id: string) => {
        setChartIds(curIds => {
            const newChartIds = [...curIds];
            if (newChartIds.includes(id)) {
                const index = newChartIds.indexOf(id);
                newChartIds.splice(index, 1);
            } else {
                newChartIds.push(id);
            }
            return newChartIds;
        });
    };

    const clearCharts = () => {
        setChartIds([]);
    };

    return {
        onToggle,
        clearCharts,
        chartIds
    };
}
