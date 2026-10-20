export namespace app {
	
	export class AboutMeFields {
	    name: string;
	    role: string;
	    interests: string;
	    objectives: string;
	    tech: string;
	    vaultUse: string;
	
	    static createFrom(source: any = {}) {
	        return new AboutMeFields(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.role = source["role"];
	        this.interests = source["interests"];
	        this.objectives = source["objectives"];
	        this.tech = source["tech"];
	        this.vaultUse = source["vaultUse"];
	    }
	}
	export class AboutMeResult {
	    ok: boolean;
	    chars: number;
	    written: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AboutMeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.chars = source["chars"];
	        this.written = source["written"];
	    }
	}
	export class ApplyQuickDetectResult {
	    applied: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplyQuickDetectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applied = source["applied"];
	    }
	}
	export class BackupValues {
	    backupFolder: string;
	    backupMax: string;
	    backupEnabled: boolean;
	    workerUrl: string;
	    vaultSealEnabled: boolean;
	    vaultSealPush: boolean;
	    vaultSealRepo: string;
	    goodReposEnabled: boolean;
	    goodReposPush: boolean;
	    goodReposRepo: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backupFolder = source["backupFolder"];
	        this.backupMax = source["backupMax"];
	        this.backupEnabled = source["backupEnabled"];
	        this.workerUrl = source["workerUrl"];
	        this.vaultSealEnabled = source["vaultSealEnabled"];
	        this.vaultSealPush = source["vaultSealPush"];
	        this.vaultSealRepo = source["vaultSealRepo"];
	        this.goodReposEnabled = source["goodReposEnabled"];
	        this.goodReposPush = source["goodReposPush"];
	        this.goodReposRepo = source["goodReposRepo"];
	    }
	}
	export class BatchResult {
	    Success: boolean;
	    Message: string;
	    Processed: number;
	    Total: number;
	    FailedLinks: number;
	    Crashed: boolean;
	    Error: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Message = source["Message"];
	        this.Processed = source["Processed"];
	        this.Total = source["Total"];
	        this.FailedLinks = source["FailedLinks"];
	        this.Crashed = source["Crashed"];
	        this.Error = source["Error"];
	    }
	}
	export class BotResolveAction {
	    allResolved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BotResolveAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allResolved = source["allResolved"];
	    }
	}
	export class HeroBox {
	    title: string;
	    body: string;
	    question?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HeroBox(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.body = source["body"];
	        this.question = source["question"];
	    }
	}
	export class BotRetryScan {
	    box?: HeroBox;
	    confirm?: HeroBox;
	    urls: string[];
	
	    static createFrom(source: any = {}) {
	        return new BotRetryScan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.box = this.convertValues(source["box"], HeroBox);
	        this.confirm = this.convertValues(source["confirm"], HeroBox);
	        this.urls = source["urls"];
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
	export class BotValues {
	    username: string;
	    token: string;
	
	    static createFrom(source: any = {}) {
	        return new BotValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.token = source["token"];
	    }
	}
	export class BotVerifyResult {
	    title: string;
	    body: string;
	    missing: string[];
	
	    static createFrom(source: any = {}) {
	        return new BotVerifyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.body = source["body"];
	        this.missing = source["missing"];
	    }
	}
	export class CredentialsValues {
	    apiId: string;
	    apiHash: string;
	    phone: string;
	    githubToken: string;
	
	    static createFrom(source: any = {}) {
	        return new CredentialsValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiId = source["apiId"];
	        this.apiHash = source["apiHash"];
	        this.phone = source["phone"];
	        this.githubToken = source["githubToken"];
	    }
	}
	
	export class HeroRender {
	    label: string;
	    glyph: string;
	    enabled: boolean;
	    danger: boolean;
	    tooltip: string;
	
	    static createFrom(source: any = {}) {
	        return new HeroRender(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.glyph = source["glyph"];
	        this.enabled = source["enabled"];
	        this.danger = source["danger"];
	        this.tooltip = source["tooltip"];
	    }
	}
	export class LLMValues {
	    provider: string;
	    numCtx: string;
	    maxOutputTokens: string;
	    ollamaUrl: string;
	    ollamaModel: string;
	    llamaCppUrl: string;
	    llamaCppKey: string;
	    llamaCppModel: string;
	    cloudUrl: string;
	    cloudKey: string;
	    cloudModel: string;
	
	    static createFrom(source: any = {}) {
	        return new LLMValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.numCtx = source["numCtx"];
	        this.maxOutputTokens = source["maxOutputTokens"];
	        this.ollamaUrl = source["ollamaUrl"];
	        this.ollamaModel = source["ollamaModel"];
	        this.llamaCppUrl = source["llamaCppUrl"];
	        this.llamaCppKey = source["llamaCppKey"];
	        this.llamaCppModel = source["llamaCppModel"];
	        this.cloudUrl = source["cloudUrl"];
	        this.cloudKey = source["cloudKey"];
	        this.cloudModel = source["cloudModel"];
	    }
	}
	export class LlamaDetectResult {
	    ok: boolean;
	    found: boolean;
	    url: string;
	    models: string[];
	    model: string;
	    saved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LlamaDetectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.found = source["found"];
	        this.url = source["url"];
	        this.models = source["models"];
	        this.model = source["model"];
	        this.saved = source["saved"];
	    }
	}
	export class OllamaModelsResult {
	    ok: boolean;
	    url: string;
	    models: string[];
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new OllamaModelsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.url = source["url"];
	        this.models = source["models"];
	        this.model = source["model"];
	    }
	}
	export class PickRow {
	    kind: string;
	    label: string;
	    icon: string;
	    path: string;
	    status: string;
	    status_text: string;
	    scannable: boolean;
	    selectable: boolean;
	    why: string;
	
	    static createFrom(source: any = {}) {
	        return new PickRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.path = source["path"];
	        this.status = source["status"];
	        this.status_text = source["status_text"];
	        this.scannable = source["scannable"];
	        this.selectable = source["selectable"];
	        this.why = source["why"];
	    }
	}
	export class PickDisplayModel {
	    title: string;
	    header: string;
	    rows: PickRow[];
	    selectable_kinds: string[];
	    any_selectable: boolean;
	    hint: string;
	
	    static createFrom(source: any = {}) {
	        return new PickDisplayModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.header = source["header"];
	        this.rows = this.convertValues(source["rows"], PickRow);
	        this.selectable_kinds = source["selectable_kinds"];
	        this.any_selectable = source["any_selectable"];
	        this.hint = source["hint"];
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
	
	export class PlanInventory {
	    total_notes: number;
	    folder_count: number;
	    root_notes: number;
	    uncategorized_notes: number;
	    trash_notes: number;
	    locked_notes: number;
	
	    static createFrom(source: any = {}) {
	        return new PlanInventory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_notes = source["total_notes"];
	        this.folder_count = source["folder_count"];
	        this.root_notes = source["root_notes"];
	        this.uncategorized_notes = source["uncategorized_notes"];
	        this.trash_notes = source["trash_notes"];
	        this.locked_notes = source["locked_notes"];
	    }
	}
	export class PlanRow {
	    id: string;
	    text: string;
	    full: string;
	    low_confidence: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlanRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	        this.full = source["full"];
	        this.low_confidence = source["low_confidence"];
	    }
	}
	export class PlanSection {
	    key: string;
	    card: string;
	    tone: string;
	    heading: string;
	    rows: PlanRow[];
	    total: number;
	    more: number;
	
	    static createFrom(source: any = {}) {
	        return new PlanSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.card = source["card"];
	        this.tone = source["tone"];
	        this.heading = source["heading"];
	        this.rows = this.convertValues(source["rows"], PlanRow);
	        this.total = source["total"];
	        this.more = source["more"];
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
	export class PlanDisplayModel {
	    title: string;
	    header: string;
	    library_line: string;
	    trash_line: string;
	    locked_line: string;
	    kept_line: string;
	    low_confidence_line: string;
	    sections: PlanSection[];
	    summary: string;
	    vault_label: string;
	    inventory: PlanInventory;
	    has_deletions: boolean;
	    has_filing: boolean;
	    has_consolidations: boolean;
	    has_low_confidence: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlanDisplayModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.header = source["header"];
	        this.library_line = source["library_line"];
	        this.trash_line = source["trash_line"];
	        this.locked_line = source["locked_line"];
	        this.kept_line = source["kept_line"];
	        this.low_confidence_line = source["low_confidence_line"];
	        this.sections = this.convertValues(source["sections"], PlanSection);
	        this.summary = source["summary"];
	        this.vault_label = source["vault_label"];
	        this.inventory = this.convertValues(source["inventory"], PlanInventory);
	        this.has_deletions = source["has_deletions"];
	        this.has_filing = source["has_filing"];
	        this.has_consolidations = source["has_consolidations"];
	        this.has_low_confidence = source["has_low_confidence"];
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
	
	
	
	export class ProxyValues {
	    enabled: boolean;
	    type: string;
	    host: string;
	    port: string;
	    useForWeb: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProxyValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.useForWeb = source["useForWeb"];
	    }
	}
	export class QuarantineValues {
	    strikeThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new QuarantineValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strikeThreshold = source["strikeThreshold"];
	    }
	}
	export class QueueBadgeState {
	    text: string;
	    pending: boolean;
	    visible: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QueueBadgeState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.pending = source["pending"];
	        this.visible = source["visible"];
	    }
	}
	export class QuickDetectResult {
	    ok: boolean;
	    found: boolean;
	    provider: string;
	    baseUrl: string;
	    models: string[];
	    detail: string;
	    propsModel: string;
	    ready?: boolean;
	    via: string;
	    applied: boolean;
	    choice: string;
	
	    static createFrom(source: any = {}) {
	        return new QuickDetectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.found = source["found"];
	        this.provider = source["provider"];
	        this.baseUrl = source["baseUrl"];
	        this.models = source["models"];
	        this.detail = source["detail"];
	        this.propsModel = source["propsModel"];
	        this.ready = source["ready"];
	        this.via = source["via"];
	        this.applied = source["applied"];
	        this.choice = source["choice"];
	    }
	}
	export class RadarResult {
	    Success: boolean;
	    Applied: number;
	    Results: Record<string, any>;
	    Lines: string[];
	    Error: string;
	
	    static createFrom(source: any = {}) {
	        return new RadarResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Applied = source["Applied"];
	        this.Results = source["Results"];
	        this.Lines = source["Lines"];
	        this.Error = source["Error"];
	    }
	}
	export class RecategorizeChange {
	    path: string;
	    oldCategory: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new RecategorizeChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.oldCategory = source["oldCategory"];
	        this.category = source["category"];
	    }
	}
	export class RecategorizeRow {
	    path: string;
	    name: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new RecategorizeRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.category = source["category"];
	    }
	}
	export class RecategorizeModel {
	    boxTitle: string;
	    boxBody: string;
	    notes: RecategorizeRow[];
	    categories: string[];
	
	    static createFrom(source: any = {}) {
	        return new RecategorizeModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.boxTitle = source["boxTitle"];
	        this.boxBody = source["boxBody"];
	        this.notes = this.convertValues(source["notes"], RecategorizeRow);
	        this.categories = source["categories"];
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
	export class RecategorizeResult {
	    moved: number;
	
	    static createFrom(source: any = {}) {
	        return new RecategorizeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.moved = source["moved"];
	    }
	}
	
	export class VaultSection {
	    label: string;
	    changed: number;
	    tip: string;
	    rows: ScoreRow[];
	
	    static createFrom(source: any = {}) {
	        return new VaultSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.changed = source["changed"];
	        this.tip = source["tip"];
	        this.rows = this.convertValues(source["rows"], ScoreRow);
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
	export class ScoreRow {
	    shown: boolean;
	    name: string;
	    value: string;
	    tone: string;
	    tip: string;
	
	    static createFrom(source: any = {}) {
	        return new ScoreRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shown = source["shown"];
	        this.name = source["name"];
	        this.value = source["value"];
	        this.tone = source["tone"];
	        this.tip = source["tip"];
	    }
	}
	export class ResultDisplayModel {
	    heading: string;
	    rows: ScoreRow[];
	    vault_sections: VaultSection[];
	    landed_total: number;
	    nothing_landed: boolean;
	    failed_moves: number;
	    nothing_line: string;
	    footnote: string;
	
	    static createFrom(source: any = {}) {
	        return new ResultDisplayModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.heading = source["heading"];
	        this.rows = this.convertValues(source["rows"], ScoreRow);
	        this.vault_sections = this.convertValues(source["vault_sections"], VaultSection);
	        this.landed_total = source["landed_total"];
	        this.nothing_landed = source["nothing_landed"];
	        this.failed_moves = source["failed_moves"];
	        this.nothing_line = source["nothing_line"];
	        this.footnote = source["footnote"];
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
	export class SaveResult {
	    saved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SaveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.saved = source["saved"];
	    }
	}
	export class ScanStory {
	    verdict: string;
	    deletions: number;
	    moves: number;
	    new_folders: number;
	    consolidations: number;
	    consolidation_notes: number;
	    applied: number;
	    moved: number;
	    folders_created: number;
	    notes_consolidated: number;
	    failed_moves: number;
	    failed: vaultscan.FailedMove[];
	    vault: string;
	    vault_kind: string;
	    vault_label: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanStory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.verdict = source["verdict"];
	        this.deletions = source["deletions"];
	        this.moves = source["moves"];
	        this.new_folders = source["new_folders"];
	        this.consolidations = source["consolidations"];
	        this.consolidation_notes = source["consolidation_notes"];
	        this.applied = source["applied"];
	        this.moved = source["moved"];
	        this.folders_created = source["folders_created"];
	        this.notes_consolidated = source["notes_consolidated"];
	        this.failed_moves = source["failed_moves"];
	        this.failed = this.convertValues(source["failed"], vaultscan.FailedMove);
	        this.vault = source["vault"];
	        this.vault_kind = source["vault_kind"];
	        this.vault_label = source["vault_label"];
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
	export class ScanResult {
	    Success: boolean;
	    Verdict: string;
	    HygieneTotal: number;
	    hygiene: Record<string, vaulthygiene.HygieneCounts>;
	    WebsitesRan: boolean;
	    ReposRan: boolean;
	    Lines: string[];
	    Crashed: boolean;
	    Error: string;
	    story?: ScanStory;
	    repos?: ScanStory;
	    vaults: string[];
	    vault: string;
	    github_vault: string;
	    vault_kind: string;
	    vault_label: string;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Verdict = source["Verdict"];
	        this.HygieneTotal = source["HygieneTotal"];
	        this.hygiene = this.convertValues(source["hygiene"], vaulthygiene.HygieneCounts, true);
	        this.WebsitesRan = source["WebsitesRan"];
	        this.ReposRan = source["ReposRan"];
	        this.Lines = source["Lines"];
	        this.Crashed = source["Crashed"];
	        this.Error = source["Error"];
	        this.story = this.convertValues(source["story"], ScanStory);
	        this.repos = this.convertValues(source["repos"], ScanStory);
	        this.vaults = source["vaults"];
	        this.vault = source["vault"];
	        this.github_vault = source["github_vault"];
	        this.vault_kind = source["vault_kind"];
	        this.vault_label = source["vault_label"];
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
	
	
	export class SealResultRow {
	    Success: boolean;
	    Message: string;
	    Level: string;
	
	    static createFrom(source: any = {}) {
	        return new SealResultRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.Message = source["Message"];
	        this.Level = source["Level"];
	    }
	}
	export class SettingsSnapshot {
	    telegramApiId: string;
	    telegramApiHash: string;
	    telegramPhone: string;
	    githubToken: string;
	    userConfigStatus: string;
	    userConfigShowable: boolean;
	    proxyEnabled: boolean;
	    proxyType: string;
	    proxyHost: string;
	    proxyPort: string;
	    proxyUseForWeb: boolean;
	    vaultPath: string;
	    vaultChoices: string[];
	    websiteVaultPath: string;
	    websiteRepoName: string;
	    manualVaultPath: string;
	    manualRepoName: string;
	    webBlocked: string;
	    webSelf: string;
	    pipelineGithub: boolean;
	    pipelineWebsites: boolean;
	    llmProvider: string;
	    llmNumCtx: string;
	    llmMaxOutputTokens: string;
	    ollamaUrl: string;
	    ollamaModel: string;
	    llamaCppUrl: string;
	    llamaCppKey: string;
	    llamaCppModel: string;
	    cloudUrl: string;
	    cloudKey: string;
	    cloudModel: string;
	    soundEnabled: boolean;
	    soundVolume: number;
	    blockedLawDomains: string[];
	    botUsername: string;
	    botToken: string;
	    strikeThreshold: number;
	    backupFolder: string;
	    backupMax: string;
	    backupEnabled: boolean;
	    vaultSealEnabled: boolean;
	    vaultSealPush: boolean;
	    vaultSealRepo: string;
	    goodReposEnabled: boolean;
	    goodReposPush: boolean;
	    goodReposRepo: string;
	    workerUrl: string;
	    hasVaultPath: boolean;
	    hasGithubToken: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingsSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.telegramApiId = source["telegramApiId"];
	        this.telegramApiHash = source["telegramApiHash"];
	        this.telegramPhone = source["telegramPhone"];
	        this.githubToken = source["githubToken"];
	        this.userConfigStatus = source["userConfigStatus"];
	        this.userConfigShowable = source["userConfigShowable"];
	        this.proxyEnabled = source["proxyEnabled"];
	        this.proxyType = source["proxyType"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUseForWeb = source["proxyUseForWeb"];
	        this.vaultPath = source["vaultPath"];
	        this.vaultChoices = source["vaultChoices"];
	        this.websiteVaultPath = source["websiteVaultPath"];
	        this.websiteRepoName = source["websiteRepoName"];
	        this.manualVaultPath = source["manualVaultPath"];
	        this.manualRepoName = source["manualRepoName"];
	        this.webBlocked = source["webBlocked"];
	        this.webSelf = source["webSelf"];
	        this.pipelineGithub = source["pipelineGithub"];
	        this.pipelineWebsites = source["pipelineWebsites"];
	        this.llmProvider = source["llmProvider"];
	        this.llmNumCtx = source["llmNumCtx"];
	        this.llmMaxOutputTokens = source["llmMaxOutputTokens"];
	        this.ollamaUrl = source["ollamaUrl"];
	        this.ollamaModel = source["ollamaModel"];
	        this.llamaCppUrl = source["llamaCppUrl"];
	        this.llamaCppKey = source["llamaCppKey"];
	        this.llamaCppModel = source["llamaCppModel"];
	        this.cloudUrl = source["cloudUrl"];
	        this.cloudKey = source["cloudKey"];
	        this.cloudModel = source["cloudModel"];
	        this.soundEnabled = source["soundEnabled"];
	        this.soundVolume = source["soundVolume"];
	        this.blockedLawDomains = source["blockedLawDomains"];
	        this.botUsername = source["botUsername"];
	        this.botToken = source["botToken"];
	        this.strikeThreshold = source["strikeThreshold"];
	        this.backupFolder = source["backupFolder"];
	        this.backupMax = source["backupMax"];
	        this.backupEnabled = source["backupEnabled"];
	        this.vaultSealEnabled = source["vaultSealEnabled"];
	        this.vaultSealPush = source["vaultSealPush"];
	        this.vaultSealRepo = source["vaultSealRepo"];
	        this.goodReposEnabled = source["goodReposEnabled"];
	        this.goodReposPush = source["goodReposPush"];
	        this.goodReposRepo = source["goodReposRepo"];
	        this.workerUrl = source["workerUrl"];
	        this.hasVaultPath = source["hasVaultPath"];
	        this.hasGithubToken = source["hasGithubToken"];
	    }
	}
	export class SoundValues {
	    enabled: boolean;
	    volume: number;
	
	    static createFrom(source: any = {}) {
	        return new SoundValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.volume = source["volume"];
	    }
	}
	export class VaultValues {
	    vaultPath: string;
	    history: string[];
	    websiteVaultPath: string;
	    websiteRepoName: string;
	    manualVaultPath: string;
	    manualRepoName: string;
	    webBlocked: string;
	    webSelf: string;
	    pipelineGithub?: boolean;
	    pipelineWebsites?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VaultValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vaultPath = source["vaultPath"];
	        this.history = source["history"];
	        this.websiteVaultPath = source["websiteVaultPath"];
	        this.websiteRepoName = source["websiteRepoName"];
	        this.manualVaultPath = source["manualVaultPath"];
	        this.manualRepoName = source["manualRepoName"];
	        this.webBlocked = source["webBlocked"];
	        this.webSelf = source["webSelf"];
	        this.pipelineGithub = source["pipelineGithub"];
	        this.pipelineWebsites = source["pipelineWebsites"];
	    }
	}
	export class SettingsValues {
	    credentials?: CredentialsValues;
	    proxy?: ProxyValues;
	    vault?: VaultValues;
	    llm?: LLMValues;
	    sound?: SoundValues;
	
	    static createFrom(source: any = {}) {
	        return new SettingsValues(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.credentials = this.convertValues(source["credentials"], CredentialsValues);
	        this.proxy = this.convertValues(source["proxy"], ProxyValues);
	        this.vault = this.convertValues(source["vault"], VaultValues);
	        this.llm = this.convertValues(source["llm"], LLMValues);
	        this.sound = this.convertValues(source["sound"], SoundValues);
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
	export class SoundPreviewResult {
	    play: boolean;
	    volume: number;
	    dataUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new SoundPreviewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.play = source["play"];
	        this.volume = source["volume"];
	        this.dataUrl = source["dataUrl"];
	    }
	}
	
	export class SourcesResult {
	    stage: string;
	    error: string;
	    urls: string[];
	    body: string;
	    notice: string;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new SourcesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = source["stage"];
	        this.error = source["error"];
	        this.urls = source["urls"];
	        this.body = source["body"];
	        this.notice = source["notice"];
	        this.level = source["level"];
	    }
	}
	export class StartServerResult {
	    ok: boolean;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new StartServerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.pid = source["pid"];
	    }
	}
	export class StatusDot {
	    text: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new StatusDot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.state = source["state"];
	    }
	}
	export class SyncBackupRepoRow {
	    kind: string;
	    label: string;
	    repo_name: string;
	    vault_path: string;
	    // Go type: storage
	    status?: any;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncBackupRepoRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.repo_name = source["repo_name"];
	        this.vault_path = source["vault_path"];
	        this.status = this.convertValues(source["status"], null);
	        this.error = source["error"];
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
	export class SyncBotFacts {
	    state: string;
	    message: string;
	    pending: any[];
	    total?: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncBotFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.message = source["message"];
	        this.pending = source["pending"];
	        this.total = source["total"];
	    }
	}
	export class SyncStat {
	    value: string;
	    label: string;
	    tone: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.label = source["label"];
	        this.tone = source["tone"];
	        this.note = source["note"];
	    }
	}
	export class SyncHero {
	    value: string;
	    label: string;
	    tone: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncHero(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.label = source["label"];
	        this.tone = source["tone"];
	    }
	}
	export class SyncCard {
	    key: string;
	    icon: string;
	    name: string;
	    role: string;
	    state: string;
	    dot_tone: string;
	    status_text: string;
	    lines: string[];
	    hint: string;
	    hero: SyncHero;
	    stats: SyncStat[];
	    foot: string[];
	    pending: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.icon = source["icon"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.state = source["state"];
	        this.dot_tone = source["dot_tone"];
	        this.status_text = source["status_text"];
	        this.lines = source["lines"];
	        this.hint = source["hint"];
	        this.hero = this.convertValues(source["hero"], SyncHero);
	        this.stats = this.convertValues(source["stats"], SyncStat);
	        this.foot = source["foot"];
	        this.pending = source["pending"];
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
	export class SyncDifferencesPayload {
	    stage?: syncreview.StageRecord;
	    heading: string;
	    rows: syncreview.LedgerRow[];
	    log_count: number;
	    found: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SyncDifferencesPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = this.convertValues(source["stage"], syncreview.StageRecord);
	        this.heading = source["heading"];
	        this.rows = this.convertValues(source["rows"], syncreview.LedgerRow);
	        this.log_count = source["log_count"];
	        this.found = source["found"];
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
	export class SyncDisplayModel {
	    title: string;
	    cards: SyncCard[];
	    all_synced: boolean;
	    calm_line: string;
	    legend: string;
	    waiting_total: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncDisplayModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.cards = this.convertValues(source["cards"], SyncCard);
	        this.all_synced = source["all_synced"];
	        this.calm_line = source["calm_line"];
	        this.legend = source["legend"];
	        this.waiting_total = source["waiting_total"];
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
	export class SyncPublishedFacts {
	    state: string;
	    sites?: number;
	    generated_at: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncPublishedFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.sites = source["sites"];
	        this.generated_at = source["generated_at"];
	        this.message = source["message"];
	    }
	}
	export class SyncWebsiteVaultSide {
	    state: string;
	    public_ready?: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncWebsiteVaultSide(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.public_ready = source["public_ready"];
	        this.message = source["message"];
	    }
	}
	export class SyncWebsiteFacts {
	    vault: SyncWebsiteVaultSide;
	    published: SyncPublishedFacts;
	
	    static createFrom(source: any = {}) {
	        return new SyncWebsiteFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vault = this.convertValues(source["vault"], SyncWebsiteVaultSide);
	        this.published = this.convertValues(source["published"], SyncPublishedFacts);
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
	export class SyncGitHubFacts {
	    rows: SyncBackupRepoRow[];
	
	    static createFrom(source: any = {}) {
	        return new SyncGitHubFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], SyncBackupRepoRow);
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
	export class SyncLastBatch {
	    when: string;
	    done: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncLastBatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.when = source["when"];
	        this.done = source["done"];
	        this.total = source["total"];
	    }
	}
	export class SyncVaultRow {
	    kind: string;
	    label: string;
	    icon: string;
	    path: string;
	    status: string;
	    note_count?: number;
	    scannable: boolean;
	    selectable: boolean;
	    why: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncVaultRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.path = source["path"];
	        this.status = source["status"];
	        this.note_count = source["note_count"];
	        this.scannable = source["scannable"];
	        this.selectable = source["selectable"];
	        this.why = source["why"];
	    }
	}
	export class SyncVaultsFacts {
	    vaults: SyncVaultRow[];
	    last_batch?: SyncLastBatch;
	
	    static createFrom(source: any = {}) {
	        return new SyncVaultsFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vaults = this.convertValues(source["vaults"], SyncVaultRow);
	        this.last_batch = this.convertValues(source["last_batch"], SyncLastBatch);
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
	export class SyncFacts {
	    bot?: SyncBotFacts;
	    vaults?: SyncVaultsFacts;
	    github?: SyncGitHubFacts;
	    website?: SyncWebsiteFacts;
	
	    static createFrom(source: any = {}) {
	        return new SyncFacts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bot = this.convertValues(source["bot"], SyncBotFacts);
	        this.vaults = this.convertValues(source["vaults"], SyncVaultsFacts);
	        this.github = this.convertValues(source["github"], SyncGitHubFacts);
	        this.website = this.convertValues(source["website"], SyncWebsiteFacts);
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
	
	
	
	
	export class SyncSnapshot {
	    stages: Record<string, syncreview.StageRecord>;
	    generated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stages = this.convertValues(source["stages"], syncreview.StageRecord, true);
	        this.generated_at = source["generated_at"];
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
	export class SyncStageDef {
	    key: string;
	    stream: string;
	    stage: string;
	    source_label: string;
	    target_label: string;
	    fix_label: string;
	    fix_action: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncStageDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.stream = source["stream"];
	        this.stage = source["stage"];
	        this.source_label = source["source_label"];
	        this.target_label = source["target_label"];
	        this.fix_label = source["fix_label"];
	        this.fix_action = source["fix_action"];
	    }
	}
	export class SyncStageTables {
	    defs: SyncStageDef[];
	    layout: Record<string, Array<string>>;
	    streams: string[];
	    columns: string[];
	
	    static createFrom(source: any = {}) {
	        return new SyncStageTables(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defs = this.convertValues(source["defs"], SyncStageDef);
	        this.layout = source["layout"];
	        this.streams = source["streams"];
	        this.columns = source["columns"];
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
	
	
	
	export class SyncViewPayload {
	    facts?: SyncFacts;
	    model?: SyncDisplayModel;
	    snapshot?: SyncSnapshot;
	    banner: syncreview.Verdict;
	    measured_at: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncViewPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.facts = this.convertValues(source["facts"], SyncFacts);
	        this.model = this.convertValues(source["model"], SyncDisplayModel);
	        this.snapshot = this.convertValues(source["snapshot"], SyncSnapshot);
	        this.banner = this.convertValues(source["banner"], syncreview.Verdict);
	        this.measured_at = source["measured_at"];
	        this.error = source["error"];
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
	
	
	export class TokenTestResult {
	    ok: boolean;
	    level: string;
	    login: string;
	
	    static createFrom(source: any = {}) {
	        return new TokenTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.level = source["level"];
	        this.login = source["login"];
	    }
	}
	export class UndoBatchResult {
	    stage: string;
	    message: string;
	    files: string[];
	    deleted: number;
	    errors: number;
	    confirmMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new UndoBatchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = source["stage"];
	        this.message = source["message"];
	        this.files = source["files"];
	        this.deleted = source["deleted"];
	        this.errors = source["errors"];
	        this.confirmMessage = source["confirmMessage"];
	    }
	}
	
	export class VaultStatus {
	    text: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new VaultStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.state = source["state"];
	    }
	}

}

export namespace main {
	
	export class BackupLadders {
	    backup: app.StatusDot;
	    vaultSeal: app.StatusDot;
	    goodRepos: app.StatusDot;
	
	    static createFrom(source: any = {}) {
	        return new BackupLadders(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backup = this.convertValues(source["backup"], app.StatusDot);
	        this.vaultSeal = this.convertValues(source["vaultSeal"], app.StatusDot);
	        this.goodRepos = this.convertValues(source["goodRepos"], app.StatusDot);
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
	export class CCRowResult {
	    icon: string;
	    title: string;
	    verdict: string;
	    items: string[];
	
	    static createFrom(source: any = {}) {
	        return new CCRowResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.icon = source["icon"];
	        this.title = source["title"];
	        this.verdict = source["verdict"];
	        this.items = source["items"];
	    }
	}
	export class CCTestResult {
	    refusal: string;
	    rows: CCRowResult[];
	    all_ok: boolean;
	    headline: string;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new CCTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refusal = source["refusal"];
	        this.rows = this.convertValues(source["rows"], CCRowResult);
	        this.all_ok = source["all_ok"];
	        this.headline = source["headline"];
	        this.level = source["level"];
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
	export class ChromeBootPayload {
	    dark_mode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ChromeBootPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dark_mode = source["dark_mode"];
	    }
	}
	export class DashboardViewResult {
	    noVault: boolean;
	    text: string;
	    quarantine: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardViewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.noVault = source["noVault"];
	        this.text = source["text"];
	        this.quarantine = source["quarantine"];
	    }
	}
	export class HeroClickResult {
	    state: string;
	    render: app.HeroRender;
	    format: string;
	    count: string;
	    box?: app.HeroBox;
	    offline: boolean;
	    confirm?: app.HeroBox;
	
	    static createFrom(source: any = {}) {
	        return new HeroClickResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.render = this.convertValues(source["render"], app.HeroRender);
	        this.format = source["format"];
	        this.count = source["count"];
	        this.box = this.convertValues(source["box"], app.HeroBox);
	        this.offline = source["offline"];
	        this.confirm = this.convertValues(source["confirm"], app.HeroBox);
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
	export class HeroStatePayload {
	    state: string;
	    render: app.HeroRender;
	    format: string;
	    count: string;
	
	    static createFrom(source: any = {}) {
	        return new HeroStatePayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.render = this.convertValues(source["render"], app.HeroRender);
	        this.format = source["format"];
	        this.count = source["count"];
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
	export class MarkResult {
	    marked: number;
	    line: string;
	
	    static createFrom(source: any = {}) {
	        return new MarkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.marked = source["marked"];
	        this.line = source["line"];
	    }
	}
	export class MoreDoorResult {
	    title: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new MoreDoorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.body = source["body"];
	    }
	}
	export class QueueCheckResult {
	    refusal: string;
	    error: string;
	    report: string;
	    badge?: app.QueueBadgeState;
	    pendingUrls: string[];
	    pendingWebsites: string[];
	    duplicates: number;
	    rawCount: number;
	
	    static createFrom(source: any = {}) {
	        return new QueueCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refusal = source["refusal"];
	        this.error = source["error"];
	        this.report = source["report"];
	        this.badge = this.convertValues(source["badge"], app.QueueBadgeState);
	        this.pendingUrls = source["pendingUrls"];
	        this.pendingWebsites = source["pendingWebsites"];
	        this.duplicates = source["duplicates"];
	        this.rawCount = source["rawCount"];
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
	export class ReopenResult {
	    ok: boolean;
	    line: string;
	
	    static createFrom(source: any = {}) {
	        return new ReopenResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.line = source["line"];
	    }
	}
	export class ScanAskPayload {
	    model?: app.PlanDisplayModel;
	    timeout_s: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanAskPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = this.convertValues(source["model"], app.PlanDisplayModel);
	        this.timeout_s = source["timeout_s"];
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
	export class ScanFlowResult {
	    started: boolean;
	    refusal: string;
	    result?: app.ScanResult;
	    scorecard?: app.ResultDisplayModel;
	
	    static createFrom(source: any = {}) {
	        return new ScanFlowResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.started = source["started"];
	        this.refusal = source["refusal"];
	        this.result = this.convertValues(source["result"], app.ScanResult);
	        this.scorecard = this.convertValues(source["scorecard"], app.ResultDisplayModel);
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
	export class SyncQueueCheckResult {
	    refusal: string;
	    error: string;
	    report: string;
	
	    static createFrom(source: any = {}) {
	        return new SyncQueueCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refusal = source["refusal"];
	        this.error = source["error"];
	        this.report = source["report"];
	    }
	}

}

export namespace syncreview {
	
	export class GapItem {
	    id: string;
	    name: string;
	    url?: string;
	    added_at: string;
	    note_path: string;
	    added_days_ago?: number;
	
	    static createFrom(source: any = {}) {
	        return new GapItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.added_at = source["added_at"];
	        this.note_path = source["note_path"];
	        this.added_days_ago = source["added_days_ago"];
	    }
	}
	export class LedgerRow {
	    stage: string;
	    item_id: string;
	    display_name: string;
	    reason: string;
	    note: string;
	    status: string;
	    date_reviewed: string;
	    date_resolved: string;
	    reopen_count: number;
	
	    static createFrom(source: any = {}) {
	        return new LedgerRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = source["stage"];
	        this.item_id = source["item_id"];
	        this.display_name = source["display_name"];
	        this.reason = source["reason"];
	        this.note = source["note"];
	        this.status = source["status"];
	        this.date_reviewed = source["date_reviewed"];
	        this.date_resolved = source["date_resolved"];
	        this.reopen_count = source["reopen_count"];
	    }
	}
	export class StageRecord {
	    key: string;
	    stream: string;
	    stage: string;
	    state: string;
	    set_up: boolean;
	    set_up_hint: string;
	    reason: string;
	    needs: string;
	    source_label: string;
	    target_label: string;
	    fix_label: string;
	    fix_action: string;
	    display_number?: number;
	    source_count?: number;
	    target_count?: number;
	    gap_count: number;
	    unexplained_count: number;
	    deferred_gap_count: number;
	    unexplained_ids: string[];
	    deferred_gap_ids: string[];
	    items: GapItem[];
	
	    static createFrom(source: any = {}) {
	        return new StageRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.stream = source["stream"];
	        this.stage = source["stage"];
	        this.state = source["state"];
	        this.set_up = source["set_up"];
	        this.set_up_hint = source["set_up_hint"];
	        this.reason = source["reason"];
	        this.needs = source["needs"];
	        this.source_label = source["source_label"];
	        this.target_label = source["target_label"];
	        this.fix_label = source["fix_label"];
	        this.fix_action = source["fix_action"];
	        this.display_number = source["display_number"];
	        this.source_count = source["source_count"];
	        this.target_count = source["target_count"];
	        this.gap_count = source["gap_count"];
	        this.unexplained_count = source["unexplained_count"];
	        this.deferred_gap_count = source["deferred_gap_count"];
	        this.unexplained_ids = source["unexplained_ids"];
	        this.deferred_gap_ids = source["deferred_gap_ids"];
	        this.items = this.convertValues(source["items"], GapItem);
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
	export class Verdict {
	    tone: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Verdict(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tone = source["tone"];
	        this.text = source["text"];
	    }
	}

}

export namespace vaulthygiene {
	
	export class HygieneCounts {
	    RemovedIndex: number;
	    RetiredStubs: number;
	    RetiredReports: number;
	    StubNames: string[];
	    RetiredLinked: number;
	    RetiredAppOwned: number;
	
	    static createFrom(source: any = {}) {
	        return new HygieneCounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RemovedIndex = source["RemovedIndex"];
	        this.RetiredStubs = source["RetiredStubs"];
	        this.RetiredReports = source["RetiredReports"];
	        this.StubNames = source["StubNames"];
	        this.RetiredLinked = source["RetiredLinked"];
	        this.RetiredAppOwned = source["RetiredAppOwned"];
	    }
	}

}

export namespace vaultscan {
	
	export class FailedMove {
	    Note: string;
	    To: string;
	    Why: string;
	
	    static createFrom(source: any = {}) {
	        return new FailedMove(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Note = source["Note"];
	        this.To = source["To"];
	        this.Why = source["Why"];
	    }
	}

}

