export namespace config {
	
	export class Config {
	    device_id: string;
	    device_name: string;
	    username: string;
	    download_dir: string;
	    tcp_listen_port: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.username = source["username"];
	        this.download_dir = source["download_dir"];
	        this.tcp_listen_port = source["tcp_listen_port"];
	    }
	}

}

export namespace models {
	
	export class ChatMessage {
	    id: string;
	    type: string;
	    content: string;
	    sender_id: string;
	    sender_name: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.sender_id = source["sender_id"];
	        this.sender_name = source["sender_name"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class Device {
	    device_id: string;
	    device_name: string;
	    username: string;
	    ip: string;
	    port: number;
	    platform: string;
	    status: number;
	    // Go type: time
	    last_seen: any;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device_id = source["device_id"];
	        this.device_name = source["device_name"];
	        this.username = source["username"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.platform = source["platform"];
	        this.status = source["status"];
	        this.last_seen = this.convertValues(source["last_seen"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class FileTransfer {
	    transfer_id: string;
	    peer_id: string;
	    direction: number;
	    status: number;
	    file_name: string;
	    file_size: number;
	    file_hash: string;
	    local_path: string;
	    transferred: number;
	    resume_offset: number;
	    chunk_size: number;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    completed_at: any;
	
	    static createFrom(source: any = {}) {
	        return new FileTransfer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transfer_id = source["transfer_id"];
	        this.peer_id = source["peer_id"];
	        this.direction = source["direction"];
	        this.status = source["status"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.file_hash = source["file_hash"];
	        this.local_path = source["local_path"];
	        this.transferred = source["transferred"];
	        this.resume_offset = source["resume_offset"];
	        this.chunk_size = source["chunk_size"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

