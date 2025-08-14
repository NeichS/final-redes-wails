export namespace server {
	
	export class FileServerInfo {
	    Address: string;
	    TCP: boolean;
	    Paths: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Address = source["Address"];
	        this.TCP = source["TCP"];
	        this.Paths = source["Paths"];
	    }
	}

}

