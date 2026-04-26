export namespace main {
	
	export class AutomationStepPayload {
	    action: string;
	    target: string;
	    value: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new AutomationStepPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.target = source["target"];
	        this.value = source["value"];
	        this.label = source["label"];
	    }
	}
	export class AutomationPayload {
	    id: string;
	    platformId: string;
	    name: string;
	    description: string;
	    steps: AutomationStepPayload[];
	    createdAt: string;
	    lastRun: string;
	    runCount: number;
	
	    static createFrom(source: any = {}) {
	        return new AutomationPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.platformId = source["platformId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.steps = this.convertValues(source["steps"], AutomationStepPayload);
	        this.createdAt = source["createdAt"];
	        this.lastRun = source["lastRun"];
	        this.runCount = source["runCount"];
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
	
	export class BrowserConfigResponse {
	    chromiumPath: string;
	    cdpPort: number;
	    headless: boolean;
	    userDataDir: string;
	    windowWidth: number;
	    windowHeight: number;
	    extraFlags: string[];
	
	    static createFrom(source: any = {}) {
	        return new BrowserConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chromiumPath = source["chromiumPath"];
	        this.cdpPort = source["cdpPort"];
	        this.headless = source["headless"];
	        this.userDataDir = source["userDataDir"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.extraFlags = source["extraFlags"];
	    }
	}
	export class BrowserStatusResponse {
	    status: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowserStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.error = source["error"];
	    }
	}
	export class CreateScheduleRequest {
	    name: string;
	    type: string;
	    platforms: string[];
	    cronExpr: string;
	    filePath?: string;
	    caption?: string;
	    hashtags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateScheduleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.platforms = source["platforms"];
	        this.cronExpr = source["cronExpr"];
	        this.filePath = source["filePath"];
	        this.caption = source["caption"];
	        this.hashtags = source["hashtags"];
	    }
	}
	export class LogEntryResponse {
	    timestamp: string;
	    level: string;
	    source: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntryResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.source = source["source"];
	        this.message = source["message"];
	    }
	}
	export class PlatformResponse {
	    id: string;
	    name: string;
	    url: string;
	    loggedIn: boolean;
	    username: string;
	    lastLogin: string;
	    lastCheck: string;
	    sessionDir: string;
	
	    static createFrom(source: any = {}) {
	        return new PlatformResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.loggedIn = source["loggedIn"];
	        this.username = source["username"];
	        this.lastLogin = source["lastLogin"];
	        this.lastCheck = source["lastCheck"];
	        this.sessionDir = source["sessionDir"];
	    }
	}
	export class RunAutomationResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new RunAutomationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class ScheduleResponse {
	    id: string;
	    name: string;
	    type: string;
	    platforms: string[];
	    cronExpr: string;
	    nextRun: string;
	    lastRun: string;
	    status: string;
	    createdAt: string;
	    runCount: number;
	    lastResult: string;
	    auto: boolean;
	    filePath?: string;
	    caption?: string;
	    hashtags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ScheduleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.platforms = source["platforms"];
	        this.cronExpr = source["cronExpr"];
	        this.nextRun = source["nextRun"];
	        this.lastRun = source["lastRun"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.runCount = source["runCount"];
	        this.lastResult = source["lastResult"];
	        this.auto = source["auto"];
	        this.filePath = source["filePath"];
	        this.caption = source["caption"];
	        this.hashtags = source["hashtags"];
	    }
	}
	export class UploadResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UploadResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}

}

export namespace workflow {
	
	export class UploadRequest {
	    platforms: string[];
	    filePath: string;
	    caption: string;
	    hashtags: string[];
	
	    static createFrom(source: any = {}) {
	        return new UploadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platforms = source["platforms"];
	        this.filePath = source["filePath"];
	        this.caption = source["caption"];
	        this.hashtags = source["hashtags"];
	    }
	}

}

