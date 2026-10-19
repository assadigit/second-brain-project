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
	        this.Crashed = source["Crashed"];
	        this.Error = source["Error"];
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

}

