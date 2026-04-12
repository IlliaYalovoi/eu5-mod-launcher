export namespace domain {
	
	export class Constraint {
	    type?: string;
	    from?: string;
	    to?: string;
	    modId?: string;
	
	    static createFrom(source: any = {}) {
	        return new Constraint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.modId = source["modId"];
	    }
	}

}

export namespace game {
	
	export class DetectedGame {
	    id: string;
	    name: string;
	    iconKey: string;
	    detected: boolean;
	    installDir: string;
	    documentsDir: string;
	    needsManualSetup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DetectedGame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.iconKey = source["iconKey"];
	        this.detected = source["detected"];
	        this.installDir = source["installDir"];
	        this.documentsDir = source["documentsDir"];
	        this.needsManualSetup = source["needsManualSetup"];
	    }
	}

}

export namespace launcher {
	
	export class LauncherCategory {
	    id: string;
	    name: string;
	    modIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new LauncherCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.modIds = source["modIds"];
	    }
	}
	export class LauncherLayout {
	    ungrouped: string[];
	    categories: LauncherCategory[];
	    order?: string[];
	    collapsed?: Record<string, boolean>;
	
	    static createFrom(source: any = {}) {
	        return new LauncherLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ungrouped = source["ungrouped"];
	        this.categories = this.convertValues(source["categories"], LauncherCategory);
	        this.order = source["order"];
	        this.collapsed = source["collapsed"];
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
	export class ModsDirStatus {
	    effectiveDir: string;
	    autoDetectedDir: string;
	    customDir: string;
	    usingCustomDir: boolean;
	    autoDetectedExists: boolean;
	    effectiveExists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModsDirStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.effectiveDir = source["effectiveDir"];
	        this.autoDetectedDir = source["autoDetectedDir"];
	        this.customDir = source["customDir"];
	        this.usingCustomDir = source["usingCustomDir"];
	        this.autoDetectedExists = source["autoDetectedExists"];
	        this.effectiveExists = source["effectiveExists"];
	    }
	}

}

export namespace mods {
	
	export class Mod {
	    id: string;
	    name: string;
	    version: string;
	    tags: string[];
	    description: string;
	    thumbnailPath: string;
	    dirPath: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Mod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.tags = source["tags"];
	        this.description = source["description"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.dirPath = source["dirPath"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace steam {
	
	export class WorkshopItem {
	    itemId: string;
	    title: string;
	    description: string;
	    previewUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkshopItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.previewUrl = source["previewUrl"];
	    }
	}

}

