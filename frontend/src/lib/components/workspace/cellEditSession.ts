export type CellEditSource = 'formula' | 'grid';

export type CellEditSession = {
	source: CellEditSource;
	sheetName: string;
	cellRef: string;
	value: string;
};
