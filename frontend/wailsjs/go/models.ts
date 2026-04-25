export namespace main {
	
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

