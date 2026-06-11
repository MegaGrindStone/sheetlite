export namespace main {
	
	export class AppearanceState {
	    mode: string;
	    systemTheme: string;
	    effectiveTheme: string;
	
	    static createFrom(source: any = {}) {
	        return new AppearanceState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.systemTheme = source["systemTheme"];
	        this.effectiveTheme = source["effectiveTheme"];
	    }
	}
	export class AppStatus {
	    kind: string;
	    message: string;
	    busy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.message = source["message"];
	        this.busy = source["busy"];
	    }
	}
	export class ScrollPosition {
	    topRow: number;
	    leftColumn: number;
	
	    static createFrom(source: any = {}) {
	        return new ScrollPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topRow = source["topRow"];
	        this.leftColumn = source["leftColumn"];
	    }
	}
	export class WorkbookViewState {
	    activeSheetName: string;
	    activeCell: CellAddress;
	    selection: CellRange;
	    zoomPercent: number;
	    scroll: ScrollPosition;
	
	    static createFrom(source: any = {}) {
	        return new WorkbookViewState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activeSheetName = source["activeSheetName"];
	        this.activeCell = this.convertValues(source["activeCell"], CellAddress);
	        this.selection = this.convertValues(source["selection"], CellRange);
	        this.zoomPercent = source["zoomPercent"];
	        this.scroll = this.convertValues(source["scroll"], ScrollPosition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CellRenderStyle {
	    textColor: string;
	    textAdjusted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CellRenderStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.textColor = source["textColor"];
	        this.textAdjusted = source["textAdjusted"];
	    }
	}
	export class CellBorderStyle {
	    side: string;
	    style: number;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new CellBorderStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.side = source["side"];
	        this.style = source["style"];
	        this.color = source["color"];
	    }
	}
	export class CellAlignmentStyle {
	    horizontal: string;
	    vertical: string;
	    wrapText: boolean;
	    textRotation: number;
	
	    static createFrom(source: any = {}) {
	        return new CellAlignmentStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.horizontal = source["horizontal"];
	        this.vertical = source["vertical"];
	        this.wrapText = source["wrapText"];
	        this.textRotation = source["textRotation"];
	    }
	}
	export class CellFillStyle {
	    type: string;
	    pattern: number;
	    color: string;
	    colors: string[];
	
	    static createFrom(source: any = {}) {
	        return new CellFillStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.pattern = source["pattern"];
	        this.color = source["color"];
	        this.colors = source["colors"];
	    }
	}
	export class CellFontStyle {
	    family: string;
	    size: number;
	    bold: boolean;
	    italic: boolean;
	    underline: string;
	    strikethrough: boolean;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new CellFontStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.family = source["family"];
	        this.size = source["size"];
	        this.bold = source["bold"];
	        this.italic = source["italic"];
	        this.underline = source["underline"];
	        this.strikethrough = source["strikethrough"];
	        this.color = source["color"];
	    }
	}
	export class CellStyle {
	    id: number;
	    numberFormatId: number;
	    numberFormat: string;
	    font: CellFontStyle;
	    fill: CellFillStyle;
	    alignment: CellAlignmentStyle;
	    borders: CellBorderStyle[];
	    render: CellRenderStyle;
	
	    static createFrom(source: any = {}) {
	        return new CellStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.numberFormatId = source["numberFormatId"];
	        this.numberFormat = source["numberFormat"];
	        this.font = this.convertValues(source["font"], CellFontStyle);
	        this.fill = this.convertValues(source["fill"], CellFillStyle);
	        this.alignment = this.convertValues(source["alignment"], CellAlignmentStyle);
	        this.borders = this.convertValues(source["borders"], CellBorderStyle);
	        this.render = this.convertValues(source["render"], CellRenderStyle);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RowLayout {
	    index: number;
	    height: number;
	    hidden: boolean;
	    outlineLevel: number;
	
	    static createFrom(source: any = {}) {
	        return new RowLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.height = source["height"];
	        this.hidden = source["hidden"];
	        this.outlineLevel = source["outlineLevel"];
	    }
	}
	export class ColumnLayout {
	    index: number;
	    name: string;
	    width: number;
	    hidden: boolean;
	    outlineLevel: number;
	    styleId: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.width = source["width"];
	        this.hidden = source["hidden"];
	        this.outlineLevel = source["outlineLevel"];
	        this.styleId = source["styleId"];
	    }
	}
	export class MergedCellRange {
	    range: CellRange;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new MergedCellRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.range = this.convertValues(source["range"], CellRange);
	        this.value = source["value"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CellData {
	    ref: string;
	    row: number;
	    column: number;
	    value: string;
	    rawValue: string;
	    formula: string;
	    hasFormula: boolean;
	    kind: string;
	    styleId: number;
	
	    static createFrom(source: any = {}) {
	        return new CellData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ref = source["ref"];
	        this.row = source["row"];
	        this.column = source["column"];
	        this.value = source["value"];
	        this.rawValue = source["rawValue"];
	        this.formula = source["formula"];
	        this.hasFormula = source["hasFormula"];
	        this.kind = source["kind"];
	        this.styleId = source["styleId"];
	    }
	}
	export class CellAddress {
	    ref: string;
	    row: number;
	    column: number;
	
	    static createFrom(source: any = {}) {
	        return new CellAddress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ref = source["ref"];
	        this.row = source["row"];
	        this.column = source["column"];
	    }
	}
	export class CellRange {
	    ref: string;
	    start: CellAddress;
	    end: CellAddress;
	
	    static createFrom(source: any = {}) {
	        return new CellRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ref = source["ref"];
	        this.start = this.convertValues(source["start"], CellAddress);
	        this.end = this.convertValues(source["end"], CellAddress);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkbookSheet {
	    index: number;
	    name: string;
	    state: string;
	    visible: boolean;
	    bounds: CellRange;
	    defaultColumnWidth: number;
	    defaultRowHeight: number;
	    cells: CellData[];
	    mergedCells: MergedCellRange[];
	    columns: ColumnLayout[];
	    rows: RowLayout[];
	
	    static createFrom(source: any = {}) {
	        return new WorkbookSheet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.state = source["state"];
	        this.visible = source["visible"];
	        this.bounds = this.convertValues(source["bounds"], CellRange);
	        this.defaultColumnWidth = source["defaultColumnWidth"];
	        this.defaultRowHeight = source["defaultRowHeight"];
	        this.cells = this.convertValues(source["cells"], CellData);
	        this.mergedCells = this.convertValues(source["mergedCells"], MergedCellRange);
	        this.columns = this.convertValues(source["columns"], ColumnLayout);
	        this.rows = this.convertValues(source["rows"], RowLayout);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkbookState {
	    hasWorkbook: boolean;
	    title: string;
	    filePath: string;
	    fileName: string;
	    sheets: WorkbookSheet[];
	    styles: CellStyle[];
	
	    static createFrom(source: any = {}) {
	        return new WorkbookState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasWorkbook = source["hasWorkbook"];
	        this.title = source["title"];
	        this.filePath = source["filePath"];
	        this.fileName = source["fileName"];
	        this.sheets = this.convertValues(source["sheets"], WorkbookSheet);
	        this.styles = this.convertValues(source["styles"], CellStyle);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppState {
	    workbook: WorkbookState;
	    view: WorkbookViewState;
	    status: AppStatus;
	    appearance: AppearanceState;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workbook = this.convertValues(source["workbook"], WorkbookState);
	        this.view = this.convertValues(source["view"], WorkbookViewState);
	        this.status = this.convertValues(source["status"], AppStatus);
	        this.appearance = this.convertValues(source["appearance"], AppearanceState);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	

}

