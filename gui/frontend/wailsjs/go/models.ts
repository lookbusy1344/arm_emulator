export namespace service {
	
	export class BreakpointInfo {
	    address: number;
	    enabled: boolean;
	    condition: string;
	
	    static createFrom(source: any = {}) {
	        return new BreakpointInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.enabled = source["enabled"];
	        this.condition = source["condition"];
	    }
	}
	export class CPSRState {
	    N: boolean;
	    Z: boolean;
	    C: boolean;
	    V: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CPSRState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.N = source["N"];
	        this.Z = source["Z"];
	        this.C = source["C"];
	        this.V = source["V"];
	    }
	}
	export class DisassemblyLine {
	    address: number;
	    opcode: number;
	    mnemonic: string;
	    symbol: string;
	
	    static createFrom(source: any = {}) {
	        return new DisassemblyLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.opcode = source["opcode"];
	        this.mnemonic = source["mnemonic"];
	        this.symbol = source["symbol"];
	    }
	}
	export class RegisterState {
	    Registers: number[];
	    CPSR: CPSRState;
	    PC: number;
	    Cycles: number;
	
	    static createFrom(source: any = {}) {
	        return new RegisterState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Registers = source["Registers"];
	        this.CPSR = this.convertValues(source["CPSR"], CPSRState);
	        this.PC = source["PC"];
	        this.Cycles = source["Cycles"];
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
	export class StackEntry {
	    address: number;
	    value: number;
	    symbol: string;
	
	    static createFrom(source: any = {}) {
	        return new StackEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.value = source["value"];
	        this.symbol = source["symbol"];
	    }
	}
	export class WatchpointInfo {
	    id: number;
	    address: number;
	    type: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WatchpointInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	    }
	}

}

