export namespace service {
	
	export class BreakpointInfo {
	    Address: number;
	    Enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BreakpointInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.Enabled = source["Enabled"];
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

}

