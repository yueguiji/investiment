export namespace asset {
	
	export class AssetCategory {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    type: string;
	    icon: string;
	    description: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new AssetCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	        this.sortOrder = source["sortOrder"];
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
	export class Asset {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    categoryId: number;
	    category: AssetCategory;
	    name: string;
	    type: string;
	    amount: number;
	    currency: string;
	    rate: number;
	    // Go type: time
	    startDate?: any;
	    // Go type: time
	    endDate?: any;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.categoryId = source["categoryId"];
	        this.category = this.convertValues(source["category"], AssetCategory);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.currency = source["currency"];
	        this.rate = source["rate"];
	        this.startDate = this.convertValues(source["startDate"], null);
	        this.endDate = this.convertValues(source["endDate"], null);
	        this.remark = source["remark"];
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
	
	export class CategorySummary {
	    categoryId: number;
	    categoryName: string;
	    type: string;
	    totalAmount: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new CategorySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.categoryId = source["categoryId"];
	        this.categoryName = source["categoryName"];
	        this.type = source["type"];
	        this.totalAmount = source["totalAmount"];
	        this.count = source["count"];
	    }
	}
	export class AssetSummary {
	    totalLiquid: number;
	    totalFixed: number;
	    totalLiability: number;
	    netAsset: number;
	    investValue: number;
	    categories: CategorySummary[];
	
	    static createFrom(source: any = {}) {
	        return new AssetSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalLiquid = source["totalLiquid"];
	        this.totalFixed = source["totalFixed"];
	        this.totalLiability = source["totalLiability"];
	        this.netAsset = source["netAsset"];
	        this.investValue = source["investValue"];
	        this.categories = this.convertValues(source["categories"], CategorySummary);
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
	
	export class HouseholdAIAnalysis {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    triggerSource: string;
	    region: string;
	    benchmarkVersion: string;
	    aiConfigId: number;
	    promptTemplateId: number;
	    modelName: string;
	    status: string;
	    prompt: string;
	    inputPayload: string;
	    analysisMarkdown: string;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdAIAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.triggerSource = source["triggerSource"];
	        this.region = source["region"];
	        this.benchmarkVersion = source["benchmarkVersion"];
	        this.aiConfigId = source["aiConfigId"];
	        this.promptTemplateId = source["promptTemplateId"];
	        this.modelName = source["modelName"];
	        this.status = source["status"];
	        this.prompt = source["prompt"];
	        this.inputPayload = source["inputPayload"];
	        this.analysisMarkdown = source["analysisMarkdown"];
	        this.errorMessage = source["errorMessage"];
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
	export class HouseholdAccount {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    accountType: string;
	    provider: string;
	    owner: string;
	    currency: string;
	    balance: number;
	    isLiquid: boolean;
	    // Go type: time
	    lastUpdatedAt?: any;
	    remark: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdAccount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.accountType = source["accountType"];
	        this.provider = source["provider"];
	        this.owner = source["owner"];
	        this.currency = source["currency"];
	        this.balance = source["balance"];
	        this.isLiquid = source["isLiquid"];
	        this.lastUpdatedAt = this.convertValues(source["lastUpdatedAt"], null);
	        this.remark = source["remark"];
	        this.isActive = source["isActive"];
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
	export class HouseholdBenchmarkRecord {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    scope: string;
	    region: string;
	    category: string;
	    value: number;
	    unit: string;
	    year: number;
	    version: string;
	    description: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdBenchmarkRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.scope = source["scope"];
	        this.region = source["region"];
	        this.category = source["category"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.year = source["year"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.isActive = source["isActive"];
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
	export class HouseholdChatMessage {
	    role: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	    }
	}
	export class HouseholdDashboardSummary {
	    totalAssets: number;
	    totalLiquidAssets: number;
	    totalFixedAssets: number;
	    totalProtection: number;
	    totalLiabilities: number;
	    netAssets: number;
	    debtRatio: number;
	    monthlyIncome: number;
	    monthlyNetIncome: number;
	    monthlyIncomeTax: number;
	    monthlyPretaxCosts: number;
	    monthlyHousingFundInflows: number;
	    monthlyDebtPayment: number;
	    monthlyEffectiveDebtPayment: number;
	    monthlyCoverageRate: number;
	    monthlyEffectiveCoverageRate: number;
	    accountCount: number;
	    fixedAssetCount: number;
	    incomeCount: number;
	    protectionCount: number;
	    liabilityCount: number;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdDashboardSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalAssets = source["totalAssets"];
	        this.totalLiquidAssets = source["totalLiquidAssets"];
	        this.totalFixedAssets = source["totalFixedAssets"];
	        this.totalProtection = source["totalProtection"];
	        this.totalLiabilities = source["totalLiabilities"];
	        this.netAssets = source["netAssets"];
	        this.debtRatio = source["debtRatio"];
	        this.monthlyIncome = source["monthlyIncome"];
	        this.monthlyNetIncome = source["monthlyNetIncome"];
	        this.monthlyIncomeTax = source["monthlyIncomeTax"];
	        this.monthlyPretaxCosts = source["monthlyPretaxCosts"];
	        this.monthlyHousingFundInflows = source["monthlyHousingFundInflows"];
	        this.monthlyDebtPayment = source["monthlyDebtPayment"];
	        this.monthlyEffectiveDebtPayment = source["monthlyEffectiveDebtPayment"];
	        this.monthlyCoverageRate = source["monthlyCoverageRate"];
	        this.monthlyEffectiveCoverageRate = source["monthlyEffectiveCoverageRate"];
	        this.accountCount = source["accountCount"];
	        this.fixedAssetCount = source["fixedAssetCount"];
	        this.incomeCount = source["incomeCount"];
	        this.protectionCount = source["protectionCount"];
	        this.liabilityCount = source["liabilityCount"];
	    }
	}
	export class HouseholdFixedAsset {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    assetType: string;
	    owner: string;
	    ownershipRatio: number;
	    currentValue: number;
	    costBasis: number;
	    location: string;
	    referenceCode: string;
	    // Go type: time
	    purchasedAt?: any;
	    // Go type: time
	    valuationDate?: any;
	    remark: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdFixedAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.assetType = source["assetType"];
	        this.owner = source["owner"];
	        this.ownershipRatio = source["ownershipRatio"];
	        this.currentValue = source["currentValue"];
	        this.costBasis = source["costBasis"];
	        this.location = source["location"];
	        this.referenceCode = source["referenceCode"];
	        this.purchasedAt = this.convertValues(source["purchasedAt"], null);
	        this.valuationDate = this.convertValues(source["valuationDate"], null);
	        this.remark = source["remark"];
	        this.isActive = source["isActive"];
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
	export class HouseholdIncome {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    incomeType: string;
	    owner: string;
	    employer: string;
	    frequency: string;
	    monthlyAmount: number;
	    annualAmount: number;
	    monthlyPersonalInsuranceContribution: number;
	    monthlyEmployerInsuranceContribution: number;
	    monthlyPersonalHousingFundContribution: number;
	    monthlyEmployerHousingFundContribution: number;
	    // Go type: time
	    lastReceivedAt?: any;
	    remark: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdIncome(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.incomeType = source["incomeType"];
	        this.owner = source["owner"];
	        this.employer = source["employer"];
	        this.frequency = source["frequency"];
	        this.monthlyAmount = source["monthlyAmount"];
	        this.annualAmount = source["annualAmount"];
	        this.monthlyPersonalInsuranceContribution = source["monthlyPersonalInsuranceContribution"];
	        this.monthlyEmployerInsuranceContribution = source["monthlyEmployerInsuranceContribution"];
	        this.monthlyPersonalHousingFundContribution = source["monthlyPersonalHousingFundContribution"];
	        this.monthlyEmployerHousingFundContribution = source["monthlyEmployerHousingFundContribution"];
	        this.lastReceivedAt = this.convertValues(source["lastReceivedAt"], null);
	        this.remark = source["remark"];
	        this.isActive = source["isActive"];
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
	export class HouseholdLiability {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    liabilityType: string;
	    lender: string;
	    owner: string;
	    currency: string;
	    principal: number;
	    outstandingPrincipal: number;
	    annualRate: number;
	    loanTermMonths: number;
	    repaymentMethod: string;
	    monthlyPayment: number;
	    extraMonthlyPayment: number;
	    // Go type: time
	    startDate?: any;
	    // Go type: time
	    firstPaymentDate?: any;
	    // Go type: time
	    maturityDate?: any;
	    autoAmortize: boolean;
	    remark: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdLiability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.liabilityType = source["liabilityType"];
	        this.lender = source["lender"];
	        this.owner = source["owner"];
	        this.currency = source["currency"];
	        this.principal = source["principal"];
	        this.outstandingPrincipal = source["outstandingPrincipal"];
	        this.annualRate = source["annualRate"];
	        this.loanTermMonths = source["loanTermMonths"];
	        this.repaymentMethod = source["repaymentMethod"];
	        this.monthlyPayment = source["monthlyPayment"];
	        this.extraMonthlyPayment = source["extraMonthlyPayment"];
	        this.startDate = this.convertValues(source["startDate"], null);
	        this.firstPaymentDate = this.convertValues(source["firstPaymentDate"], null);
	        this.maturityDate = this.convertValues(source["maturityDate"], null);
	        this.autoAmortize = source["autoAmortize"];
	        this.remark = source["remark"];
	        this.isActive = source["isActive"];
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
	export class HouseholdLiabilitySchedule {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    liabilityId: number;
	    // Go type: time
	    dueDate?: any;
	    periodNumber: number;
	    openingPrincipal: number;
	    principalPaid: number;
	    interestPaid: number;
	    paymentAmount: number;
	    closingPrincipal: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdLiabilitySchedule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.liabilityId = source["liabilityId"];
	        this.dueDate = this.convertValues(source["dueDate"], null);
	        this.periodNumber = source["periodNumber"];
	        this.openingPrincipal = source["openingPrincipal"];
	        this.principalPaid = source["principalPaid"];
	        this.interestPaid = source["interestPaid"];
	        this.paymentAmount = source["paymentAmount"];
	        this.closingPrincipal = source["closingPrincipal"];
	        this.status = source["status"];
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
	export class HouseholdLiabilityTrendPoint {
	    month: string;
	    totalOutstanding: number;
	    totalPayment: number;
	    principalPaid: number;
	    interestPaid: number;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdLiabilityTrendPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.totalOutstanding = source["totalOutstanding"];
	        this.totalPayment = source["totalPayment"];
	        this.principalPaid = source["principalPaid"];
	        this.interestPaid = source["interestPaid"];
	    }
	}
	export class HouseholdLiquidAssetDistributionItem {
	    name: string;
	    accountType: string;
	    provider: string;
	    owner: string;
	    balance: number;
	    shareOfLiquid: number;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdLiquidAssetDistributionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.accountType = source["accountType"];
	        this.provider = source["provider"];
	        this.owner = source["owner"];
	        this.balance = source["balance"];
	        this.shareOfLiquid = source["shareOfLiquid"];
	    }
	}
	export class HouseholdLiquidAssetTrendPoint {
	    date: string;
	    totalLiquidAssets: number;
	    monthlyNetIncome: number;
	    monthlyDebtPayment: number;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdLiquidAssetTrendPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.totalLiquidAssets = source["totalLiquidAssets"];
	        this.monthlyNetIncome = source["monthlyNetIncome"];
	        this.monthlyDebtPayment = source["monthlyDebtPayment"];
	    }
	}
	export class HouseholdMember {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    relationship: string;
	    gender: string;
	    // Go type: time
	    birthDate?: any;
	    occupation: string;
	    city: string;
	    annualIncome: number;
	    notes: string;
	    isPrimary: boolean;
	    isDependent: boolean;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdMember(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.relationship = source["relationship"];
	        this.gender = source["gender"];
	        this.birthDate = this.convertValues(source["birthDate"], null);
	        this.occupation = source["occupation"];
	        this.city = source["city"];
	        this.annualIncome = source["annualIncome"];
	        this.notes = source["notes"];
	        this.isPrimary = source["isPrimary"];
	        this.isDependent = source["isDependent"];
	        this.isActive = source["isActive"];
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
	export class HouseholdProfile {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    householdName: string;
	    region: string;
	    cityTier: string;
	    membersCount: number;
	    dependentsCount: number;
	    housingStatus: string;
	    riskPreference: string;
	    monthlyHouseholdSpend: number;
	    annualHouseholdSpend: number;
	    primaryIncomeSource: string;
	    monthlyPersonalInsuranceContribution: number;
	    monthlyHousingFundContribution: number;
	    monthlyOtherPretaxDeduction: number;
	    monthlyChildcareDeduction: number;
	    monthlyHousingLoanDeduction: number;
	    monthlyElderlyCareDeduction: number;
	    monthlyOtherSpecialDeduction: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.householdName = source["householdName"];
	        this.region = source["region"];
	        this.cityTier = source["cityTier"];
	        this.membersCount = source["membersCount"];
	        this.dependentsCount = source["dependentsCount"];
	        this.housingStatus = source["housingStatus"];
	        this.riskPreference = source["riskPreference"];
	        this.monthlyHouseholdSpend = source["monthlyHouseholdSpend"];
	        this.annualHouseholdSpend = source["annualHouseholdSpend"];
	        this.primaryIncomeSource = source["primaryIncomeSource"];
	        this.monthlyPersonalInsuranceContribution = source["monthlyPersonalInsuranceContribution"];
	        this.monthlyHousingFundContribution = source["monthlyHousingFundContribution"];
	        this.monthlyOtherPretaxDeduction = source["monthlyOtherPretaxDeduction"];
	        this.monthlyChildcareDeduction = source["monthlyChildcareDeduction"];
	        this.monthlyHousingLoanDeduction = source["monthlyHousingLoanDeduction"];
	        this.monthlyElderlyCareDeduction = source["monthlyElderlyCareDeduction"];
	        this.monthlyOtherSpecialDeduction = source["monthlyOtherSpecialDeduction"];
	        this.notes = source["notes"];
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
	export class HouseholdProtection {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    protectionType: string;
	    sourceIncomeId: number;
	    owner: string;
	    provider: string;
	    employer: string;
	    currentBalance: number;
	    monthlyPersonalContribution: number;
	    monthlyEmployerContribution: number;
	    monthlyPremium: number;
	    coverageAmount: number;
	    // Go type: time
	    nextDueDate?: any;
	    remark: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdProtection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.protectionType = source["protectionType"];
	        this.sourceIncomeId = source["sourceIncomeId"];
	        this.owner = source["owner"];
	        this.provider = source["provider"];
	        this.employer = source["employer"];
	        this.currentBalance = source["currentBalance"];
	        this.monthlyPersonalContribution = source["monthlyPersonalContribution"];
	        this.monthlyEmployerContribution = source["monthlyEmployerContribution"];
	        this.monthlyPremium = source["monthlyPremium"];
	        this.coverageAmount = source["coverageAmount"];
	        this.nextDueDate = this.convertValues(source["nextDueDate"], null);
	        this.remark = source["remark"];
	        this.isActive = source["isActive"];
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
	export class HouseholdSnapshot {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    // Go type: time
	    snapshotDate?: any;
	    totalAssets: number;
	    totalLiquidAssets: number;
	    totalFixedAssets: number;
	    totalProtection: number;
	    totalLiabilities: number;
	    netAssets: number;
	    debtRatio: number;
	    monthlyIncome: number;
	    monthlyNetIncome: number;
	    monthlyIncomeTax: number;
	    monthlyPretaxCosts: number;
	    monthlyHousingFundInflows: number;
	    monthlyDebtPayment: number;
	    monthlyEffectiveDebtPayment: number;
	    monthlyCoverageRate: number;
	    monthlyEffectiveCoverageRate: number;
	    triggerSource: string;
	
	    static createFrom(source: any = {}) {
	        return new HouseholdSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.snapshotDate = this.convertValues(source["snapshotDate"], null);
	        this.totalAssets = source["totalAssets"];
	        this.totalLiquidAssets = source["totalLiquidAssets"];
	        this.totalFixedAssets = source["totalFixedAssets"];
	        this.totalProtection = source["totalProtection"];
	        this.totalLiabilities = source["totalLiabilities"];
	        this.netAssets = source["netAssets"];
	        this.debtRatio = source["debtRatio"];
	        this.monthlyIncome = source["monthlyIncome"];
	        this.monthlyNetIncome = source["monthlyNetIncome"];
	        this.monthlyIncomeTax = source["monthlyIncomeTax"];
	        this.monthlyPretaxCosts = source["monthlyPretaxCosts"];
	        this.monthlyHousingFundInflows = source["monthlyHousingFundInflows"];
	        this.monthlyDebtPayment = source["monthlyDebtPayment"];
	        this.monthlyEffectiveDebtPayment = source["monthlyEffectiveDebtPayment"];
	        this.monthlyCoverageRate = source["monthlyCoverageRate"];
	        this.monthlyEffectiveCoverageRate = source["monthlyEffectiveCoverageRate"];
	        this.triggerSource = source["triggerSource"];
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

export namespace data {
	
	export class AIConfig {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    name: string;
	    baseUrl: string;
	    apiKey: string;
	    modelName: string;
	    maxTokens: number;
	    temperature: number;
	    timeOut: number;
	    httpProxy: string;
	    httpProxyEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.name = source["name"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.modelName = source["modelName"];
	        this.maxTokens = source["maxTokens"];
	        this.temperature = source["temperature"];
	        this.timeOut = source["timeOut"];
	        this.httpProxy = source["httpProxy"];
	        this.httpProxyEnabled = source["httpProxyEnabled"];
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
	export class AllStockInfoPageData {
	    list: models.AllStockInfo[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new AllStockInfoPageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], models.AllStockInfo);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class AllStockInfoQuery {
	    page: number;
	    pageSize: number;
	    securityCode: string;
	    securityName: string;
	    market: string;
	    industry: string;
	    concept: string;
	    minPrice: string;
	    maxPrice: string;
	    minChange: string;
	    maxChange: string;
	    searchKeyWord: string;
	
	    static createFrom(source: any = {}) {
	        return new AllStockInfoQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.securityCode = source["securityCode"];
	        this.securityName = source["securityName"];
	        this.market = source["market"];
	        this.industry = source["industry"];
	        this.concept = source["concept"];
	        this.minPrice = source["minPrice"];
	        this.maxPrice = source["maxPrice"];
	        this.minChange = source["minChange"];
	        this.maxChange = source["maxChange"];
	        this.searchKeyWord = source["searchKeyWord"];
	    }
	}
	export class FundBasic {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    code: string;
	    name: string;
	    fullName: string;
	    type: string;
	    establishment: string;
	    scale: string;
	    company: string;
	    manager: string;
	    rating: string;
	    trackingTarget: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedUnitTime: string;
	    netAccumulated?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    netGrowth36?: number;
	    netGrowth60?: number;
	    netGrowthYTD?: number;
	    netGrowthAll?: number;
	    netGrowth7?: number;
	    maxDrawdown3?: number;
	    maxDrawdown6?: number;
	    maxDrawdown12?: number;
	    volatility12?: number;
	    sharpe12?: number;
	    calmar12?: number;
	    stageRank1m: number;
	    stageRank1mTotal: number;
	    stageRank3m: number;
	    stageRank3mTotal: number;
	    stageRank6m: number;
	    stageRank6mTotal: number;
	    stageRank12m: number;
	    stageRank12mTotal: number;
	    redeemFeeFreeDays: number;
	    topIndustry: string;
	    topIndustryWeight?: number;
	    topIndustryDate: string;
	    screenUpdatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FundBasic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.fullName = source["fullName"];
	        this.type = source["type"];
	        this.establishment = source["establishment"];
	        this.scale = source["scale"];
	        this.company = source["company"];
	        this.manager = source["manager"];
	        this.rating = source["rating"];
	        this.trackingTarget = source["trackingTarget"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedUnitTime = source["netEstimatedUnitTime"];
	        this.netAccumulated = source["netAccumulated"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.netGrowth36 = source["netGrowth36"];
	        this.netGrowth60 = source["netGrowth60"];
	        this.netGrowthYTD = source["netGrowthYTD"];
	        this.netGrowthAll = source["netGrowthAll"];
	        this.netGrowth7 = source["netGrowth7"];
	        this.maxDrawdown3 = source["maxDrawdown3"];
	        this.maxDrawdown6 = source["maxDrawdown6"];
	        this.maxDrawdown12 = source["maxDrawdown12"];
	        this.volatility12 = source["volatility12"];
	        this.sharpe12 = source["sharpe12"];
	        this.calmar12 = source["calmar12"];
	        this.stageRank1m = source["stageRank1m"];
	        this.stageRank1mTotal = source["stageRank1mTotal"];
	        this.stageRank3m = source["stageRank3m"];
	        this.stageRank3mTotal = source["stageRank3mTotal"];
	        this.stageRank6m = source["stageRank6m"];
	        this.stageRank6mTotal = source["stageRank6mTotal"];
	        this.stageRank12m = source["stageRank12m"];
	        this.stageRank12mTotal = source["stageRank12mTotal"];
	        this.redeemFeeFreeDays = source["redeemFeeFreeDays"];
	        this.topIndustry = source["topIndustry"];
	        this.topIndustryWeight = source["topIndustryWeight"];
	        this.topIndustryDate = source["topIndustryDate"];
	        this.screenUpdatedAt = source["screenUpdatedAt"];
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
	export class FollowedFund {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    code: string;
	    name: string;
	    isWatchlist: boolean;
	    watchGroup: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedUnitTime: string;
	    netAccumulated?: number;
	    netEstimatedRate?: number;
	    fundBasic: FundBasic;
	
	    static createFrom(source: any = {}) {
	        return new FollowedFund(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.isWatchlist = source["isWatchlist"];
	        this.watchGroup = source["watchGroup"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedUnitTime = source["netEstimatedUnitTime"];
	        this.netAccumulated = source["netAccumulated"];
	        this.netEstimatedRate = source["netEstimatedRate"];
	        this.fundBasic = this.convertValues(source["fundBasic"], FundBasic);
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
	
	export class Group {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    sort: number;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.sort = source["sort"];
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
	export class GroupStock {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    stockCode: string;
	    groupId: number;
	    groupInfo: Group;
	
	    static createFrom(source: any = {}) {
	        return new GroupStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.stockCode = source["stockCode"];
	        this.groupId = source["groupId"];
	        this.groupInfo = this.convertValues(source["groupInfo"], Group);
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
	export class SettingConfig {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    tushareToken: string;
	    localPushEnable: boolean;
	    dingPushEnable: boolean;
	    dingRobot: string;
	    updateBasicInfoOnStart: boolean;
	    refreshInterval: number;
	    openAiEnable: boolean;
	    prompt: string;
	    checkUpdate: boolean;
	    questionTemplate: string;
	    crawlTimeOut: number;
	    kDays: number;
	    enableDanmu: boolean;
	    browserPath: string;
	    enableNews: boolean;
	    darkTheme: boolean;
	    browserPoolSize: number;
	    enableFund: boolean;
	    enablePushNews: boolean;
	    enableOnlyPushRedNews: boolean;
	    sponsorCode: string;
	    httpProxy: string;
	    httpProxyEnabled: boolean;
	    enableAgent: boolean;
	    qgqpBId: string;
	    assetUnlockPassword: string;
	    aiConfigs: AIConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SettingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.tushareToken = source["tushareToken"];
	        this.localPushEnable = source["localPushEnable"];
	        this.dingPushEnable = source["dingPushEnable"];
	        this.dingRobot = source["dingRobot"];
	        this.updateBasicInfoOnStart = source["updateBasicInfoOnStart"];
	        this.refreshInterval = source["refreshInterval"];
	        this.openAiEnable = source["openAiEnable"];
	        this.prompt = source["prompt"];
	        this.checkUpdate = source["checkUpdate"];
	        this.questionTemplate = source["questionTemplate"];
	        this.crawlTimeOut = source["crawlTimeOut"];
	        this.kDays = source["kDays"];
	        this.enableDanmu = source["enableDanmu"];
	        this.browserPath = source["browserPath"];
	        this.enableNews = source["enableNews"];
	        this.darkTheme = source["darkTheme"];
	        this.browserPoolSize = source["browserPoolSize"];
	        this.enableFund = source["enableFund"];
	        this.enablePushNews = source["enablePushNews"];
	        this.enableOnlyPushRedNews = source["enableOnlyPushRedNews"];
	        this.sponsorCode = source["sponsorCode"];
	        this.httpProxy = source["httpProxy"];
	        this.httpProxyEnabled = source["httpProxyEnabled"];
	        this.enableAgent = source["enableAgent"];
	        this.qgqpBId = source["qgqpBId"];
	        this.assetUnlockPassword = source["assetUnlockPassword"];
	        this.aiConfigs = this.convertValues(source["aiConfigs"], AIConfig);
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
	export class StockBasic {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    ts_code: string;
	    symbol: string;
	    name: string;
	    area: string;
	    industry: string;
	    fullname: string;
	    enname: string;
	    cnspell: string;
	    market: string;
	    exchange: string;
	    curr_type: string;
	    list_status: string;
	    list_date: string;
	    delist_date: string;
	    is_hs: string;
	    act_name: string;
	    act_ent_type: string;
	    bk_name: string;
	    bk_code: string;
	
	    static createFrom(source: any = {}) {
	        return new StockBasic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.ts_code = source["ts_code"];
	        this.symbol = source["symbol"];
	        this.name = source["name"];
	        this.area = source["area"];
	        this.industry = source["industry"];
	        this.fullname = source["fullname"];
	        this.enname = source["enname"];
	        this.cnspell = source["cnspell"];
	        this.market = source["market"];
	        this.exchange = source["exchange"];
	        this.curr_type = source["curr_type"];
	        this.list_status = source["list_status"];
	        this.list_date = source["list_date"];
	        this.delist_date = source["delist_date"];
	        this.is_hs = source["is_hs"];
	        this.act_name = source["act_name"];
	        this.act_ent_type = source["act_ent_type"];
	        this.bk_name = source["bk_name"];
	        this.bk_code = source["bk_code"];
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
	export class StockChangeItem {
	    time: string;
	    code: string;
	    name: string;
	    market: number;
	    changeType: number;
	    typeName: string;
	    volume: number;
	    price: number;
	    changeRate: number;
	    amount: number;
	
	    static createFrom(source: any = {}) {
	        return new StockChangeItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.market = source["market"];
	        this.changeType = source["changeType"];
	        this.typeName = source["typeName"];
	        this.volume = source["volume"];
	        this.price = source["price"];
	        this.changeRate = source["changeRate"];
	        this.amount = source["amount"];
	    }
	}
	export class StockChangesResponse {
	    totalCount: number;
	    data: StockChangeItem[];
	
	    static createFrom(source: any = {}) {
	        return new StockChangesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCount = source["totalCount"];
	        this.data = this.convertValues(source["data"], StockChangeItem);
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
	export class StockInfo {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    "日期": string;
	    "时间": string;
	    "股票代码": string;
	    "股票名称": string;
	    "上次当前价格": number;
	    "当前价格": string;
	    "成交的股票数": string;
	    "成交金额": string;
	    "今日开盘价": string;
	    "昨日收盘价": string;
	    "今日最高价": string;
	    "今日最低价": string;
	    "竞买价": string;
	    "竞卖价": string;
	    "买一报价": string;
	    "买一申报": string;
	    "买二报价": string;
	    "买二申报": string;
	    "买三报价": string;
	    "买三申报": string;
	    "买四报价": string;
	    "买四申报": string;
	    "买五报价": string;
	    "买五申报": string;
	    "卖一报价": string;
	    "卖一申报": string;
	    "卖二报价": string;
	    "卖二申报": string;
	    "卖三报价": string;
	    "卖三申报": string;
	    "卖四报价": string;
	    "卖四申报": string;
	    "卖五报价": string;
	    "卖五申报": string;
	    "市场": string;
	    "盘前盘后": string;
	    "盘前盘后涨跌幅": string;
	    changePercent: number;
	    changePrice: number;
	    highRate: number;
	    lowRate: number;
	    costPrice: number;
	    costVolume: number;
	    profit: number;
	    profitAmount: number;
	    profitAmountToday: number;
	    sort: number;
	    alarmChangePercent: number;
	    alarmPrice: number;
	    Groups: GroupStock[];
	
	    static createFrom(source: any = {}) {
	        return new StockInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this["日期"] = source["日期"];
	        this["时间"] = source["时间"];
	        this["股票代码"] = source["股票代码"];
	        this["股票名称"] = source["股票名称"];
	        this["上次当前价格"] = source["上次当前价格"];
	        this["当前价格"] = source["当前价格"];
	        this["成交的股票数"] = source["成交的股票数"];
	        this["成交金额"] = source["成交金额"];
	        this["今日开盘价"] = source["今日开盘价"];
	        this["昨日收盘价"] = source["昨日收盘价"];
	        this["今日最高价"] = source["今日最高价"];
	        this["今日最低价"] = source["今日最低价"];
	        this["竞买价"] = source["竞买价"];
	        this["竞卖价"] = source["竞卖价"];
	        this["买一报价"] = source["买一报价"];
	        this["买一申报"] = source["买一申报"];
	        this["买二报价"] = source["买二报价"];
	        this["买二申报"] = source["买二申报"];
	        this["买三报价"] = source["买三报价"];
	        this["买三申报"] = source["买三申报"];
	        this["买四报价"] = source["买四报价"];
	        this["买四申报"] = source["买四申报"];
	        this["买五报价"] = source["买五报价"];
	        this["买五申报"] = source["买五申报"];
	        this["卖一报价"] = source["卖一报价"];
	        this["卖一申报"] = source["卖一申报"];
	        this["卖二报价"] = source["卖二报价"];
	        this["卖二申报"] = source["卖二申报"];
	        this["卖三报价"] = source["卖三报价"];
	        this["卖三申报"] = source["卖三申报"];
	        this["卖四报价"] = source["卖四报价"];
	        this["卖四申报"] = source["卖四申报"];
	        this["卖五报价"] = source["卖五报价"];
	        this["卖五申报"] = source["卖五申报"];
	        this["市场"] = source["市场"];
	        this["盘前盘后"] = source["盘前盘后"];
	        this["盘前盘后涨跌幅"] = source["盘前盘后涨跌幅"];
	        this.changePercent = source["changePercent"];
	        this.changePrice = source["changePrice"];
	        this.highRate = source["highRate"];
	        this.lowRate = source["lowRate"];
	        this.costPrice = source["costPrice"];
	        this.costVolume = source["costVolume"];
	        this.profit = source["profit"];
	        this.profitAmount = source["profitAmount"];
	        this.profitAmountToday = source["profitAmountToday"];
	        this.sort = source["sort"];
	        this.alarmChangePercent = source["alarmChangePercent"];
	        this.alarmPrice = source["alarmPrice"];
	        this.Groups = this.convertValues(source["Groups"], GroupStock);
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

export namespace models {
	
	export class AIResponseResult {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    chatId: string;
	    modelName: string;
	    stockCode: string;
	    stockName: string;
	    question: string;
	    content: string;
	    IsDel: number;
	
	    static createFrom(source: any = {}) {
	        return new AIResponseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.chatId = source["chatId"];
	        this.modelName = source["modelName"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.question = source["question"];
	        this.content = source["content"];
	        this.IsDel = source["IsDel"];
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
	export class AIResponseResultPageData {
	    list: AIResponseResult[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new AIResponseResultPageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], AIResponseResult);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class AIResponseResultQuery {
	    page: number;
	    pageSize: number;
	    chatId: string;
	    modelName: string;
	    stockCode: string;
	    stockName: string;
	    question: string;
	    startDate: string;
	    endDate: string;
	
	    static createFrom(source: any = {}) {
	        return new AIResponseResultQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.chatId = source["chatId"];
	        this.modelName = source["modelName"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.question = source["question"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	    }
	}
	export class AiRecommendStocks {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    // Go type: time
	    dataTime?: any;
	    modelName: string;
	    stockCode: string;
	    stockName: string;
	    bkCode: string;
	    bkName: string;
	    stockPrice: string;
	    stockCurrentPrice: string;
	    stockCurrentPriceTime: string;
	    stockClosePrice: string;
	    stockPrePrice: string;
	    recommendReason: string;
	    recommendBuyPrice: string;
	    recommendBuyPriceMin: number;
	    recommendBuyPriceMax: number;
	    recommendStopProfitPrice: string;
	    recommendStopProfitPriceMin: number;
	    recommendStopProfitPriceMax: number;
	    recommendStopLossPrice: string;
	    riskRemarks: string;
	    remarks: string;
	
	    static createFrom(source: any = {}) {
	        return new AiRecommendStocks(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.dataTime = this.convertValues(source["dataTime"], null);
	        this.modelName = source["modelName"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.bkCode = source["bkCode"];
	        this.bkName = source["bkName"];
	        this.stockPrice = source["stockPrice"];
	        this.stockCurrentPrice = source["stockCurrentPrice"];
	        this.stockCurrentPriceTime = source["stockCurrentPriceTime"];
	        this.stockClosePrice = source["stockClosePrice"];
	        this.stockPrePrice = source["stockPrePrice"];
	        this.recommendReason = source["recommendReason"];
	        this.recommendBuyPrice = source["recommendBuyPrice"];
	        this.recommendBuyPriceMin = source["recommendBuyPriceMin"];
	        this.recommendBuyPriceMax = source["recommendBuyPriceMax"];
	        this.recommendStopProfitPrice = source["recommendStopProfitPrice"];
	        this.recommendStopProfitPriceMin = source["recommendStopProfitPriceMin"];
	        this.recommendStopProfitPriceMax = source["recommendStopProfitPriceMax"];
	        this.recommendStopLossPrice = source["recommendStopLossPrice"];
	        this.riskRemarks = source["riskRemarks"];
	        this.remarks = source["remarks"];
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
	export class AiRecommendStocksPageData {
	    list: AiRecommendStocks[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new AiRecommendStocksPageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], AiRecommendStocks);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class AiRecommendStocksQuery {
	    page: number;
	    pageSize: number;
	    modelName: string;
	    stockCode: string;
	    stockName: string;
	    bkCode: string;
	    bkName: string;
	    startDate: string;
	    endDate: string;
	
	    static createFrom(source: any = {}) {
	        return new AiRecommendStocksQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.modelName = source["modelName"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.bkCode = source["bkCode"];
	        this.bkName = source["bkName"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	    }
	}
	export class AllStockInfo {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    SECUCODE: string;
	    SECURITY_CODE: string;
	    SECURITY_NAME_ABBR: string;
	    NEW_PRICE: string;
	    CHANGE_RATE: string;
	    VOLUME_RATIO: string;
	    HIGH_PRICE: string;
	    LOW_PRICE: string;
	    PRE_CLOSE_PRICE: string;
	    VOLUME: string;
	    DEAL_AMOUNT: string;
	    TURNOVERRATE: string;
	    MARKET: string;
	    CONCEPT: string;
	    INDUSTRY: string;
	    MAX_TRADE_DATE: string;
	
	    static createFrom(source: any = {}) {
	        return new AllStockInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.SECUCODE = source["SECUCODE"];
	        this.SECURITY_CODE = source["SECURITY_CODE"];
	        this.SECURITY_NAME_ABBR = source["SECURITY_NAME_ABBR"];
	        this.NEW_PRICE = source["NEW_PRICE"];
	        this.CHANGE_RATE = source["CHANGE_RATE"];
	        this.VOLUME_RATIO = source["VOLUME_RATIO"];
	        this.HIGH_PRICE = source["HIGH_PRICE"];
	        this.LOW_PRICE = source["LOW_PRICE"];
	        this.PRE_CLOSE_PRICE = source["PRE_CLOSE_PRICE"];
	        this.VOLUME = source["VOLUME"];
	        this.DEAL_AMOUNT = source["DEAL_AMOUNT"];
	        this.TURNOVERRATE = source["TURNOVERRATE"];
	        this.MARKET = source["MARKET"];
	        this.CONCEPT = source["CONCEPT"];
	        this.INDUSTRY = source["INDUSTRY"];
	        this.MAX_TRADE_DATE = source["MAX_TRADE_DATE"];
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
	export class AllStocksResp {
	    version: any;
	    // Go type: struct { Nextpage bool "json:\"nextpage\""; Currentpage int "json:\"currentpage\""; Data []models
	    result: any;
	    success: boolean;
	    message: string;
	    code: number;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new AllStocksResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.result = this.convertValues(source["result"], Object);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.code = source["code"];
	        this.url = source["url"];
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
	export class PromptTemplate {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    name: string;
	    content: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.name = source["name"];
	        this.content = source["content"];
	        this.type = source["type"];
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
	export class PromptTemplatePageData {
	    list: PromptTemplate[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplatePageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], PromptTemplate);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class PromptTemplateQuery {
	    page: number;
	    pageSize: number;
	    name: string;
	    type: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplateQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.content = source["content"];
	    }
	}
	export class SentimentResult {
	    Score: number;
	    Category: number;
	    PositiveCount: number;
	    NegativeCount: number;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new SentimentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Score = source["Score"];
	        this.Category = source["Category"];
	        this.PositiveCount = source["PositiveCount"];
	        this.NegativeCount = source["NegativeCount"];
	        this.Description = source["Description"];
	    }
	}
	export class StockChangeHistory {
	    id: number;
	    changeTime: string;
	    changeDate: string;
	    stockCode: string;
	    stockName: string;
	    market: number;
	    changeType: number;
	    typeName: string;
	    volume: number;
	    price: number;
	    changeRate: number;
	    amount: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new StockChangeHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.changeTime = source["changeTime"];
	        this.changeDate = source["changeDate"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.market = source["market"];
	        this.changeType = source["changeType"];
	        this.typeName = source["typeName"];
	        this.volume = source["volume"];
	        this.price = source["price"];
	        this.changeRate = source["changeRate"];
	        this.amount = source["amount"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class StockChangeHistoryPageData {
	    list: StockChangeHistory[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new StockChangeHistoryPageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], StockChangeHistory);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class StockChangeHistoryQuery {
	    stockCode: string;
	    stockName: string;
	    changeType: number;
	    changeTypes: number[];
	    typeName: string;
	    startDate: string;
	    endDate: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new StockChangeHistoryQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.changeType = source["changeType"];
	        this.changeTypes = source["changeTypes"];
	        this.typeName = source["typeName"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class StockInfo {
	    SECUCODE: string;
	    SECURITY_CODE: string;
	    SECURITY_NAME_ABBR: string;
	    NEW_PRICE: any;
	    CHANGE_RATE: any;
	    VOLUME_RATIO: any;
	    HIGH_PRICE: any;
	    LOW_PRICE: any;
	    PRE_CLOSE_PRICE: any;
	    VOLUME: any;
	    DEAL_AMOUNT: any;
	    TURNOVERRATE: any;
	    MARKET: string;
	    CONCEPT: any;
	    INDUSTRY: string;
	    MAX_TRADE_DATE: string;
	
	    static createFrom(source: any = {}) {
	        return new StockInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SECUCODE = source["SECUCODE"];
	        this.SECURITY_CODE = source["SECURITY_CODE"];
	        this.SECURITY_NAME_ABBR = source["SECURITY_NAME_ABBR"];
	        this.NEW_PRICE = source["NEW_PRICE"];
	        this.CHANGE_RATE = source["CHANGE_RATE"];
	        this.VOLUME_RATIO = source["VOLUME_RATIO"];
	        this.HIGH_PRICE = source["HIGH_PRICE"];
	        this.LOW_PRICE = source["LOW_PRICE"];
	        this.PRE_CLOSE_PRICE = source["PRE_CLOSE_PRICE"];
	        this.VOLUME = source["VOLUME"];
	        this.DEAL_AMOUNT = source["DEAL_AMOUNT"];
	        this.TURNOVERRATE = source["TURNOVERRATE"];
	        this.MARKET = source["MARKET"];
	        this.CONCEPT = source["CONCEPT"];
	        this.INDUSTRY = source["INDUSTRY"];
	        this.MAX_TRADE_DATE = source["MAX_TRADE_DATE"];
	    }
	}
	export class TechnicalIndicators {
	    MACD_GOLDEN_FORK: boolean;
	    KDJ_GOLDEN_FORK: boolean;
	    BREAK_THROUGH: boolean;
	    LOW_FUNDS_INFLOW: boolean;
	    HIGH_FUNDS_OUTFLOW: boolean;
	    BREAKUP_MA_5DAYS: boolean;
	    LONG_AVG_ARRAY: boolean;
	    SHORT_AVG_ARRAY: boolean;
	    UPPER_LARGE_VOLUME: boolean;
	    DOWN_NARROW_VOLUME: boolean;
	    ONE_DAYANG_LINE: boolean;
	    TWO_DAYANG_LINES: boolean;
	    RISE_SUN: boolean;
	    POWER_FULGUN: boolean;
	    RESTORE_JUSTICE: boolean;
	    DOWN_7DAYS: boolean;
	    UPPER_8DAYS: boolean;
	    UPPER_9DAYS: boolean;
	    UPPER_4DAYS: boolean;
	    HEAVEN_RULE: boolean;
	    UPSIDE_VOLUME: boolean;
	    BEARISH_ENGULFING: boolean;
	    REVERSING_HAMMER: boolean;
	    SHOOTING_STAR: boolean;
	    EVENING_STAR: boolean;
	    FIRST_DAWN: boolean;
	    PREGNANT: boolean;
	    BLACK_CLOUD_TOPS: boolean;
	    MORNING_STAR: boolean;
	    NARROW_FINISH: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TechnicalIndicators(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.MACD_GOLDEN_FORK = source["MACD_GOLDEN_FORK"];
	        this.KDJ_GOLDEN_FORK = source["KDJ_GOLDEN_FORK"];
	        this.BREAK_THROUGH = source["BREAK_THROUGH"];
	        this.LOW_FUNDS_INFLOW = source["LOW_FUNDS_INFLOW"];
	        this.HIGH_FUNDS_OUTFLOW = source["HIGH_FUNDS_OUTFLOW"];
	        this.BREAKUP_MA_5DAYS = source["BREAKUP_MA_5DAYS"];
	        this.LONG_AVG_ARRAY = source["LONG_AVG_ARRAY"];
	        this.SHORT_AVG_ARRAY = source["SHORT_AVG_ARRAY"];
	        this.UPPER_LARGE_VOLUME = source["UPPER_LARGE_VOLUME"];
	        this.DOWN_NARROW_VOLUME = source["DOWN_NARROW_VOLUME"];
	        this.ONE_DAYANG_LINE = source["ONE_DAYANG_LINE"];
	        this.TWO_DAYANG_LINES = source["TWO_DAYANG_LINES"];
	        this.RISE_SUN = source["RISE_SUN"];
	        this.POWER_FULGUN = source["POWER_FULGUN"];
	        this.RESTORE_JUSTICE = source["RESTORE_JUSTICE"];
	        this.DOWN_7DAYS = source["DOWN_7DAYS"];
	        this.UPPER_8DAYS = source["UPPER_8DAYS"];
	        this.UPPER_9DAYS = source["UPPER_9DAYS"];
	        this.UPPER_4DAYS = source["UPPER_4DAYS"];
	        this.HEAVEN_RULE = source["HEAVEN_RULE"];
	        this.UPSIDE_VOLUME = source["UPSIDE_VOLUME"];
	        this.BEARISH_ENGULFING = source["BEARISH_ENGULFING"];
	        this.REVERSING_HAMMER = source["REVERSING_HAMMER"];
	        this.SHOOTING_STAR = source["SHOOTING_STAR"];
	        this.EVENING_STAR = source["EVENING_STAR"];
	        this.FIRST_DAWN = source["FIRST_DAWN"];
	        this.PREGNANT = source["PREGNANT"];
	        this.BLACK_CLOUD_TOPS = source["BLACK_CLOUD_TOPS"];
	        this.MORNING_STAR = source["MORNING_STAR"];
	        this.NARROW_FINISH = source["NARROW_FINISH"];
	    }
	}
	export class VersionInfo {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    version: string;
	    content: string;
	    icon: string;
	    alipay: string;
	    wxpay: string;
	    wxgzh: string;
	    buildTimeStamp: number;
	    officialStatement: string;
	    danmuWebsocketUrl: string;
	    messageWallUrl: string;
	    assetUnlockEnabled: boolean;
	    IsDel: number;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.version = source["version"];
	        this.content = source["content"];
	        this.icon = source["icon"];
	        this.alipay = source["alipay"];
	        this.wxpay = source["wxpay"];
	        this.wxgzh = source["wxgzh"];
	        this.buildTimeStamp = source["buildTimeStamp"];
	        this.officialStatement = source["officialStatement"];
	        this.danmuWebsocketUrl = source["danmuWebsocketUrl"];
	        this.messageWallUrl = source["messageWallUrl"];
	        this.assetUnlockEnabled = source["assetUnlockEnabled"];
	        this.IsDel = source["IsDel"];
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

export namespace portfolio {
	
	export class AllocationItem {
	    label: string;
	    value: number;
	    ratio: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new AllocationItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	        this.ratio = source["ratio"];
	        this.count = source["count"];
	    }
	}
	export class BetterFundMetric {
	    key: string;
	    label: string;
	    better: string;
	    candidateValue?: number;
	    referenceValue?: number;
	    delta?: number;
	    advantage?: number;
	    weight: number;
	    contribution: number;
	    candidateRank: number;
	    referenceRank: number;
	    rankTotal: number;
	    candidatePercentile?: number;
	    referencePercentile?: number;
	
	    static createFrom(source: any = {}) {
	        return new BetterFundMetric(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.better = source["better"];
	        this.candidateValue = source["candidateValue"];
	        this.referenceValue = source["referenceValue"];
	        this.delta = source["delta"];
	        this.advantage = source["advantage"];
	        this.weight = source["weight"];
	        this.contribution = source["contribution"];
	        this.candidateRank = source["candidateRank"];
	        this.referenceRank = source["referenceRank"];
	        this.rankTotal = source["rankTotal"];
	        this.candidatePercentile = source["candidatePercentile"];
	        this.referencePercentile = source["referencePercentile"];
	    }
	}
	export class BetterFundCandidate {
	    code: string;
	    name: string;
	    fundType: string;
	    trackingTarget: string;
	    category: string;
	    categoryLabel: string;
	    riskLevel: string;
	    redeemFeeFreeDays: number;
	    company: string;
	    manager: string;
	    rating: string;
	    scale: string;
	    topIndustry: string;
	    topIndustryWeight?: number;
	    topIndustryDate: string;
	    netGrowth7?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    maxDrawdown3?: number;
	    maxDrawdown6?: number;
	    maxDrawdown12?: number;
	    volatility12?: number;
	    sharpe12?: number;
	    calmar12?: number;
	    stageRank1m: number;
	    stageRank1mTotal: number;
	    stageRank3m: number;
	    stageRank3mTotal: number;
	    stageRank6m: number;
	    stageRank6mTotal: number;
	    stageRank12m: number;
	    stageRank12mTotal: number;
	    screenUpdatedAt: string;
	    watchlist: boolean;
	    recommendationRank: number;
	    betterScore: number;
	    reasonSummary: string;
	    scopeLabel: string;
	    comparedUniverse: number;
	    reasons: string[];
	    metrics: BetterFundMetric[];
	
	    static createFrom(source: any = {}) {
	        return new BetterFundCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.fundType = source["fundType"];
	        this.trackingTarget = source["trackingTarget"];
	        this.category = source["category"];
	        this.categoryLabel = source["categoryLabel"];
	        this.riskLevel = source["riskLevel"];
	        this.redeemFeeFreeDays = source["redeemFeeFreeDays"];
	        this.company = source["company"];
	        this.manager = source["manager"];
	        this.rating = source["rating"];
	        this.scale = source["scale"];
	        this.topIndustry = source["topIndustry"];
	        this.topIndustryWeight = source["topIndustryWeight"];
	        this.topIndustryDate = source["topIndustryDate"];
	        this.netGrowth7 = source["netGrowth7"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.maxDrawdown3 = source["maxDrawdown3"];
	        this.maxDrawdown6 = source["maxDrawdown6"];
	        this.maxDrawdown12 = source["maxDrawdown12"];
	        this.volatility12 = source["volatility12"];
	        this.sharpe12 = source["sharpe12"];
	        this.calmar12 = source["calmar12"];
	        this.stageRank1m = source["stageRank1m"];
	        this.stageRank1mTotal = source["stageRank1mTotal"];
	        this.stageRank3m = source["stageRank3m"];
	        this.stageRank3mTotal = source["stageRank3mTotal"];
	        this.stageRank6m = source["stageRank6m"];
	        this.stageRank6mTotal = source["stageRank6mTotal"];
	        this.stageRank12m = source["stageRank12m"];
	        this.stageRank12mTotal = source["stageRank12mTotal"];
	        this.screenUpdatedAt = source["screenUpdatedAt"];
	        this.watchlist = source["watchlist"];
	        this.recommendationRank = source["recommendationRank"];
	        this.betterScore = source["betterScore"];
	        this.reasonSummary = source["reasonSummary"];
	        this.scopeLabel = source["scopeLabel"];
	        this.comparedUniverse = source["comparedUniverse"];
	        this.reasons = source["reasons"];
	        this.metrics = this.convertValues(source["metrics"], BetterFundMetric);
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
	
	export class BetterFundQuery {
	    referenceCode: string;
	    sameTypeOnly: boolean;
	    sameSubTypeOnly: boolean;
	    dimension: string;
	    networkRefresh: boolean;
	    feeFree7: boolean;
	    feeFree30: boolean;
	    includeAClass: boolean;
	    onlyAClass: boolean;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new BetterFundQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.referenceCode = source["referenceCode"];
	        this.sameTypeOnly = source["sameTypeOnly"];
	        this.sameSubTypeOnly = source["sameSubTypeOnly"];
	        this.dimension = source["dimension"];
	        this.networkRefresh = source["networkRefresh"];
	        this.feeFree7 = source["feeFree7"];
	        this.feeFree30 = source["feeFree30"];
	        this.includeAClass = source["includeAClass"];
	        this.onlyAClass = source["onlyAClass"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class FundRefreshStatus {
	    state: string;
	    stateLabel: string;
	    scope: string;
	    refreshing: boolean;
	    needsRefresh: boolean;
	    triggered: boolean;
	    currentDate: string;
	    lastRefreshHint: string;
	    updatedToday: number;
	    screenedCount: number;
	    universeCount: number;
	    targetCount: number;
	    targetUpdated: number;
	    targetPending: number;
	    progressCurrent: number;
	    progressTotal: number;
	    currentCode: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new FundRefreshStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.stateLabel = source["stateLabel"];
	        this.scope = source["scope"];
	        this.refreshing = source["refreshing"];
	        this.needsRefresh = source["needsRefresh"];
	        this.triggered = source["triggered"];
	        this.currentDate = source["currentDate"];
	        this.lastRefreshHint = source["lastRefreshHint"];
	        this.updatedToday = source["updatedToday"];
	        this.screenedCount = source["screenedCount"];
	        this.universeCount = source["universeCount"];
	        this.targetCount = source["targetCount"];
	        this.targetUpdated = source["targetUpdated"];
	        this.targetPending = source["targetPending"];
	        this.progressCurrent = source["progressCurrent"];
	        this.progressTotal = source["progressTotal"];
	        this.currentCode = source["currentCode"];
	        this.message = source["message"];
	    }
	}
	export class FundScreenerItem {
	    code: string;
	    name: string;
	    fundType: string;
	    trackingTarget: string;
	    category: string;
	    categoryLabel: string;
	    riskLevel: string;
	    redeemFeeFreeDays: number;
	    company: string;
	    manager: string;
	    rating: string;
	    scale: string;
	    topIndustry: string;
	    topIndustryWeight?: number;
	    topIndustryDate: string;
	    netGrowth7?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    maxDrawdown3?: number;
	    maxDrawdown6?: number;
	    maxDrawdown12?: number;
	    volatility12?: number;
	    sharpe12?: number;
	    calmar12?: number;
	    stageRank1m: number;
	    stageRank1mTotal: number;
	    stageRank3m: number;
	    stageRank3mTotal: number;
	    stageRank6m: number;
	    stageRank6mTotal: number;
	    stageRank12m: number;
	    stageRank12mTotal: number;
	    screenUpdatedAt: string;
	    watchlist: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FundScreenerItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.fundType = source["fundType"];
	        this.trackingTarget = source["trackingTarget"];
	        this.category = source["category"];
	        this.categoryLabel = source["categoryLabel"];
	        this.riskLevel = source["riskLevel"];
	        this.redeemFeeFreeDays = source["redeemFeeFreeDays"];
	        this.company = source["company"];
	        this.manager = source["manager"];
	        this.rating = source["rating"];
	        this.scale = source["scale"];
	        this.topIndustry = source["topIndustry"];
	        this.topIndustryWeight = source["topIndustryWeight"];
	        this.topIndustryDate = source["topIndustryDate"];
	        this.netGrowth7 = source["netGrowth7"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.maxDrawdown3 = source["maxDrawdown3"];
	        this.maxDrawdown6 = source["maxDrawdown6"];
	        this.maxDrawdown12 = source["maxDrawdown12"];
	        this.volatility12 = source["volatility12"];
	        this.sharpe12 = source["sharpe12"];
	        this.calmar12 = source["calmar12"];
	        this.stageRank1m = source["stageRank1m"];
	        this.stageRank1mTotal = source["stageRank1mTotal"];
	        this.stageRank3m = source["stageRank3m"];
	        this.stageRank3mTotal = source["stageRank3mTotal"];
	        this.stageRank6m = source["stageRank6m"];
	        this.stageRank6mTotal = source["stageRank6mTotal"];
	        this.stageRank12m = source["stageRank12m"];
	        this.stageRank12mTotal = source["stageRank12mTotal"];
	        this.screenUpdatedAt = source["screenUpdatedAt"];
	        this.watchlist = source["watchlist"];
	    }
	}
	export class BetterFundResult {
	    reference: FundScreenerItem;
	    candidates: BetterFundCandidate[];
	    dimension: string;
	    sortLabel: string;
	    scopeLabel: string;
	    comparedUniverse: number;
	    universeTotal: number;
	    refreshedCount: number;
	    networkRefresh: boolean;
	    fallbackApplied: boolean;
	    dataHint: string;
	    total: number;
	    page: number;
	    pageSize: number;
	    refreshStatus: FundRefreshStatus;
	
	    static createFrom(source: any = {}) {
	        return new BetterFundResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reference = this.convertValues(source["reference"], FundScreenerItem);
	        this.candidates = this.convertValues(source["candidates"], BetterFundCandidate);
	        this.dimension = source["dimension"];
	        this.sortLabel = source["sortLabel"];
	        this.scopeLabel = source["scopeLabel"];
	        this.comparedUniverse = source["comparedUniverse"];
	        this.universeTotal = source["universeTotal"];
	        this.refreshedCount = source["refreshedCount"];
	        this.networkRefresh = source["networkRefresh"];
	        this.fallbackApplied = source["fallbackApplied"];
	        this.dataHint = source["dataHint"];
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.refreshStatus = this.convertValues(source["refreshStatus"], FundRefreshStatus);
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
	export class FundCompareQuery {
	    codes: string[];
	
	    static createFrom(source: any = {}) {
	        return new FundCompareQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.codes = source["codes"];
	    }
	}
	export class FundCompareResult {
	    items: FundScreenerItem[];
	    total: number;
	    missingCodes: string[];
	    refreshedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FundCompareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], FundScreenerItem);
	        this.total = source["total"];
	        this.missingCodes = source["missingCodes"];
	        this.refreshedAt = source["refreshedAt"];
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
	export class FundEstimatePoint {
	    timestamp: number;
	    time: string;
	    estimatedUnit: number;
	    estimatedRate?: number;
	
	    static createFrom(source: any = {}) {
	        return new FundEstimatePoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.time = source["time"];
	        this.estimatedUnit = source["estimatedUnit"];
	        this.estimatedRate = source["estimatedRate"];
	    }
	}
	export class FundHoldingView {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    stockCode: string;
	    stockName: string;
	    holdingType: string;
	    market: string;
	    avgCost: number;
	    quantity: number;
	    currentPrice: number;
	    latestDailyRate?: number;
	    latestDailyUpdatedAt: string;
	    profitLoss: number;
	    profitRate: number;
	    totalCost: number;
	    totalValue: number;
	    todayChange: number;
	    todayRate: number;
	    // Go type: time
	    buyDate?: any;
	    brokerName: string;
	    accountTag: string;
	    remark: string;
	    fundType: string;
	    fundCompany: string;
	    fundManager: string;
	    fundRating: string;
	    fundScale: string;
	    category: string;
	    categoryLabel: string;
	    riskLevel: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedTime: string;
	    netEstimatedRate?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    netGrowth36?: number;
	    netGrowthYTD?: number;
	    estimateUpdated: boolean;
	    estimateStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new FundHoldingView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.holdingType = source["holdingType"];
	        this.market = source["market"];
	        this.avgCost = source["avgCost"];
	        this.quantity = source["quantity"];
	        this.currentPrice = source["currentPrice"];
	        this.latestDailyRate = source["latestDailyRate"];
	        this.latestDailyUpdatedAt = source["latestDailyUpdatedAt"];
	        this.profitLoss = source["profitLoss"];
	        this.profitRate = source["profitRate"];
	        this.totalCost = source["totalCost"];
	        this.totalValue = source["totalValue"];
	        this.todayChange = source["todayChange"];
	        this.todayRate = source["todayRate"];
	        this.buyDate = this.convertValues(source["buyDate"], null);
	        this.brokerName = source["brokerName"];
	        this.accountTag = source["accountTag"];
	        this.remark = source["remark"];
	        this.fundType = source["fundType"];
	        this.fundCompany = source["fundCompany"];
	        this.fundManager = source["fundManager"];
	        this.fundRating = source["fundRating"];
	        this.fundScale = source["fundScale"];
	        this.category = source["category"];
	        this.categoryLabel = source["categoryLabel"];
	        this.riskLevel = source["riskLevel"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedTime = source["netEstimatedTime"];
	        this.netEstimatedRate = source["netEstimatedRate"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.netGrowth36 = source["netGrowth36"];
	        this.netGrowthYTD = source["netGrowthYTD"];
	        this.estimateUpdated = source["estimateUpdated"];
	        this.estimateStatus = source["estimateStatus"];
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
	export class Holding {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    stockCode: string;
	    stockName: string;
	    holdingType: string;
	    market: string;
	    avgCost: number;
	    quantity: number;
	    currentPrice: number;
	    latestDailyRate?: number;
	    latestDailyUpdatedAt: string;
	    profitLoss: number;
	    profitRate: number;
	    totalCost: number;
	    totalValue: number;
	    todayChange: number;
	    todayRate: number;
	    // Go type: time
	    buyDate?: any;
	    brokerName: string;
	    accountTag: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new Holding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.holdingType = source["holdingType"];
	        this.market = source["market"];
	        this.avgCost = source["avgCost"];
	        this.quantity = source["quantity"];
	        this.currentPrice = source["currentPrice"];
	        this.latestDailyRate = source["latestDailyRate"];
	        this.latestDailyUpdatedAt = source["latestDailyUpdatedAt"];
	        this.profitLoss = source["profitLoss"];
	        this.profitRate = source["profitRate"];
	        this.totalCost = source["totalCost"];
	        this.totalValue = source["totalValue"];
	        this.todayChange = source["todayChange"];
	        this.todayRate = source["todayRate"];
	        this.buyDate = this.convertValues(source["buyDate"], null);
	        this.brokerName = source["brokerName"];
	        this.accountTag = source["accountTag"];
	        this.remark = source["remark"];
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
	export class PortfolioSummary {
	    totalCost: number;
	    totalValue: number;
	    totalProfit: number;
	    totalProfitRate: number;
	    todayProfit: number;
	    stockCount: number;
	    fundCount: number;
	    bondFundCount: number;
	    cashFundCount: number;
	    equityFundCount: number;
	    fundValue: number;
	    stockValue: number;
	    bondFundValue: number;
	    cashFundValue: number;
	    equityFundValue: number;
	    holdings: Holding[];
	
	    static createFrom(source: any = {}) {
	        return new PortfolioSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCost = source["totalCost"];
	        this.totalValue = source["totalValue"];
	        this.totalProfit = source["totalProfit"];
	        this.totalProfitRate = source["totalProfitRate"];
	        this.todayProfit = source["todayProfit"];
	        this.stockCount = source["stockCount"];
	        this.fundCount = source["fundCount"];
	        this.bondFundCount = source["bondFundCount"];
	        this.cashFundCount = source["cashFundCount"];
	        this.equityFundCount = source["equityFundCount"];
	        this.fundValue = source["fundValue"];
	        this.stockValue = source["stockValue"];
	        this.bondFundValue = source["bondFundValue"];
	        this.cashFundValue = source["cashFundValue"];
	        this.equityFundValue = source["equityFundValue"];
	        this.holdings = this.convertValues(source["holdings"], Holding);
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
	export class FundPortfolioDashboard {
	    summary: PortfolioSummary;
	    positions: FundHoldingView[];
	    typeAllocation: AllocationItem[];
	    platformAllocation: AllocationItem[];
	    accountAllocation: AllocationItem[];
	    companyAllocation: AllocationItem[];
	    conservativeRatio: number;
	    bondAllocationRatio: number;
	    estimatedProfitToday: number;
	
	    static createFrom(source: any = {}) {
	        return new FundPortfolioDashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = this.convertValues(source["summary"], PortfolioSummary);
	        this.positions = this.convertValues(source["positions"], FundHoldingView);
	        this.typeAllocation = this.convertValues(source["typeAllocation"], AllocationItem);
	        this.platformAllocation = this.convertValues(source["platformAllocation"], AllocationItem);
	        this.accountAllocation = this.convertValues(source["accountAllocation"], AllocationItem);
	        this.companyAllocation = this.convertValues(source["companyAllocation"], AllocationItem);
	        this.conservativeRatio = source["conservativeRatio"];
	        this.bondAllocationRatio = source["bondAllocationRatio"];
	        this.estimatedProfitToday = source["estimatedProfitToday"];
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
	export class FundPositionInput {
	    stockCode: string;
	    stockName: string;
	    positionAmount: number;
	    costAmount: number;
	    brokerName: string;
	    accountTag: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new FundPositionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.positionAmount = source["positionAmount"];
	        this.costAmount = source["costAmount"];
	        this.brokerName = source["brokerName"];
	        this.accountTag = source["accountTag"];
	        this.remark = source["remark"];
	    }
	}
	export class FundStageRanking {
	    period: string;
	    returnRate?: number;
	    similarAverageRate?: number;
	    benchmarkLabel: string;
	    benchmarkRate?: number;
	    rank: number;
	    rankTotal: number;
	    rankPercentile?: number;
	    rankDelta: number;
	    rankDeltaDirection: string;
	    quartile: string;
	
	    static createFrom(source: any = {}) {
	        return new FundStageRanking(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.period = source["period"];
	        this.returnRate = source["returnRate"];
	        this.similarAverageRate = source["similarAverageRate"];
	        this.benchmarkLabel = source["benchmarkLabel"];
	        this.benchmarkRate = source["benchmarkRate"];
	        this.rank = source["rank"];
	        this.rankTotal = source["rankTotal"];
	        this.rankPercentile = source["rankPercentile"];
	        this.rankDelta = source["rankDelta"];
	        this.rankDeltaDirection = source["rankDeltaDirection"];
	        this.quartile = source["quartile"];
	    }
	}
	export class FundTrendPoint {
	    timestamp: number;
	    date: string;
	    value: number;
	    dailyReturn?: number;
	
	    static createFrom(source: any = {}) {
	        return new FundTrendPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.date = source["date"];
	        this.value = source["value"];
	        this.dailyReturn = source["dailyReturn"];
	    }
	}
	export class FundProfile {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    stockCode: string;
	    stockName: string;
	    holdingType: string;
	    market: string;
	    avgCost: number;
	    quantity: number;
	    currentPrice: number;
	    latestDailyRate?: number;
	    latestDailyUpdatedAt: string;
	    profitLoss: number;
	    profitRate: number;
	    totalCost: number;
	    totalValue: number;
	    todayChange: number;
	    todayRate: number;
	    // Go type: time
	    buyDate?: any;
	    brokerName: string;
	    accountTag: string;
	    remark: string;
	    fundType: string;
	    fundCompany: string;
	    fundManager: string;
	    fundRating: string;
	    fundScale: string;
	    category: string;
	    categoryLabel: string;
	    riskLevel: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedTime: string;
	    netEstimatedRate?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    netGrowth36?: number;
	    netGrowthYTD?: number;
	    estimateUpdated: boolean;
	    estimateStatus: string;
	    trend: FundTrendPoint[];
	    trendUpdatedAt: string;
	    latestReturn?: number;
	    estimateTrend: FundEstimatePoint[];
	    estimateTrendUpdatedAt: string;
	    estimateLatestRate?: number;
	    stageRankings: FundStageRanking[];
	    stageRankingsUpdatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FundProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.holdingType = source["holdingType"];
	        this.market = source["market"];
	        this.avgCost = source["avgCost"];
	        this.quantity = source["quantity"];
	        this.currentPrice = source["currentPrice"];
	        this.latestDailyRate = source["latestDailyRate"];
	        this.latestDailyUpdatedAt = source["latestDailyUpdatedAt"];
	        this.profitLoss = source["profitLoss"];
	        this.profitRate = source["profitRate"];
	        this.totalCost = source["totalCost"];
	        this.totalValue = source["totalValue"];
	        this.todayChange = source["todayChange"];
	        this.todayRate = source["todayRate"];
	        this.buyDate = this.convertValues(source["buyDate"], null);
	        this.brokerName = source["brokerName"];
	        this.accountTag = source["accountTag"];
	        this.remark = source["remark"];
	        this.fundType = source["fundType"];
	        this.fundCompany = source["fundCompany"];
	        this.fundManager = source["fundManager"];
	        this.fundRating = source["fundRating"];
	        this.fundScale = source["fundScale"];
	        this.category = source["category"];
	        this.categoryLabel = source["categoryLabel"];
	        this.riskLevel = source["riskLevel"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedTime = source["netEstimatedTime"];
	        this.netEstimatedRate = source["netEstimatedRate"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.netGrowth36 = source["netGrowth36"];
	        this.netGrowthYTD = source["netGrowthYTD"];
	        this.estimateUpdated = source["estimateUpdated"];
	        this.estimateStatus = source["estimateStatus"];
	        this.trend = this.convertValues(source["trend"], FundTrendPoint);
	        this.trendUpdatedAt = source["trendUpdatedAt"];
	        this.latestReturn = source["latestReturn"];
	        this.estimateTrend = this.convertValues(source["estimateTrend"], FundEstimatePoint);
	        this.estimateTrendUpdatedAt = source["estimateTrendUpdatedAt"];
	        this.estimateLatestRate = source["estimateLatestRate"];
	        this.stageRankings = this.convertValues(source["stageRankings"], FundStageRanking);
	        this.stageRankingsUpdatedAt = source["stageRankingsUpdatedAt"];
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
	export class FundRecommendationRefreshStatus {
	    state: string;
	    stateLabel: string;
	    refreshing: boolean;
	    triggered: boolean;
	    currentDate: string;
	    watchlistCount: number;
	    completedCount: number;
	    pendingCount: number;
	    failedCount: number;
	    progressCurrent: number;
	    progressTotal: number;
	    currentCode: string;
	    lastRefreshHint: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new FundRecommendationRefreshStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.stateLabel = source["stateLabel"];
	        this.refreshing = source["refreshing"];
	        this.triggered = source["triggered"];
	        this.currentDate = source["currentDate"];
	        this.watchlistCount = source["watchlistCount"];
	        this.completedCount = source["completedCount"];
	        this.pendingCount = source["pendingCount"];
	        this.failedCount = source["failedCount"];
	        this.progressCurrent = source["progressCurrent"];
	        this.progressTotal = source["progressTotal"];
	        this.currentCode = source["currentCode"];
	        this.lastRefreshHint = source["lastRefreshHint"];
	        this.message = source["message"];
	    }
	}
	
	
	export class FundScreenerQuery {
	    keyword: string;
	    fundType: string;
	    category: string;
	    industry: string;
	    minReturn7?: number;
	    minReturn1?: number;
	    minReturn3?: number;
	    maxDrawdown12?: number;
	    onlyWatchlist: boolean;
	    page: number;
	    pageSize: number;
	    sortBy: string;
	    sortOrder: string;
	
	    static createFrom(source: any = {}) {
	        return new FundScreenerQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword = source["keyword"];
	        this.fundType = source["fundType"];
	        this.category = source["category"];
	        this.industry = source["industry"];
	        this.minReturn7 = source["minReturn7"];
	        this.minReturn1 = source["minReturn1"];
	        this.minReturn3 = source["minReturn3"];
	        this.maxDrawdown12 = source["maxDrawdown12"];
	        this.onlyWatchlist = source["onlyWatchlist"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.sortBy = source["sortBy"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class FundScreenerResult {
	    items: FundScreenerItem[];
	    total: number;
	    page: number;
	    pageSize: number;
	    universeCount: number;
	    screenedCount: number;
	    typeOptions: string[];
	    categoryOptions: string[];
	    industryOptions: string[];
	    lastRefreshHint: string;
	    refreshStatus: FundRefreshStatus;
	
	    static createFrom(source: any = {}) {
	        return new FundScreenerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], FundScreenerItem);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.universeCount = source["universeCount"];
	        this.screenedCount = source["screenedCount"];
	        this.typeOptions = source["typeOptions"];
	        this.categoryOptions = source["categoryOptions"];
	        this.industryOptions = source["industryOptions"];
	        this.lastRefreshHint = source["lastRefreshHint"];
	        this.refreshStatus = this.convertValues(source["refreshStatus"], FundRefreshStatus);
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
	
	
	
	
	export class ProfitSnapshot {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    // Go type: time
	    snapshotDate?: any;
	    totalCost: number;
	    totalValue: number;
	    totalProfit: number;
	    profitRate: number;
	    stockValue: number;
	    fundValue: number;
	
	    static createFrom(source: any = {}) {
	        return new ProfitSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.snapshotDate = this.convertValues(source["snapshotDate"], null);
	        this.totalCost = source["totalCost"];
	        this.totalValue = source["totalValue"];
	        this.totalProfit = source["totalProfit"];
	        this.profitRate = source["profitRate"];
	        this.stockValue = source["stockValue"];
	        this.fundValue = source["fundValue"];
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
	export class Transaction {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    holdingId: number;
	    stockCode: string;
	    stockName: string;
	    holdingType: string;
	    brokerName: string;
	    accountTag: string;
	    type: string;
	    price: number;
	    quantity: number;
	    amount: number;
	    fee: number;
	    // Go type: time
	    tradeDate?: any;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.holdingId = source["holdingId"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.holdingType = source["holdingType"];
	        this.brokerName = source["brokerName"];
	        this.accountTag = source["accountTag"];
	        this.type = source["type"];
	        this.price = source["price"];
	        this.quantity = source["quantity"];
	        this.amount = source["amount"];
	        this.fee = source["fee"];
	        this.tradeDate = this.convertValues(source["tradeDate"], null);
	        this.remark = source["remark"];
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

export namespace quant {
	
	export class GenerateRequest {
	    strategyDescription: string;
	    brokerPlatform: string;
	    strategyType: string;
	    scriptCategory: string;
	    stockCodes: string;
	    riskLevel: string;
	    capital: number;
	    factorTags: string;
	    sceneTags: string;
	    aiModel: string;
	    promptTemplateId: number;
	    baseCode: string;
	    existingScriptName: string;
	    existingDescription: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategyDescription = source["strategyDescription"];
	        this.brokerPlatform = source["brokerPlatform"];
	        this.strategyType = source["strategyType"];
	        this.scriptCategory = source["scriptCategory"];
	        this.stockCodes = source["stockCodes"];
	        this.riskLevel = source["riskLevel"];
	        this.capital = source["capital"];
	        this.factorTags = source["factorTags"];
	        this.sceneTags = source["sceneTags"];
	        this.aiModel = source["aiModel"];
	        this.promptTemplateId = source["promptTemplateId"];
	        this.baseCode = source["baseCode"];
	        this.existingScriptName = source["existingScriptName"];
	        this.existingDescription = source["existingDescription"];
	    }
	}
	export class TemplateCategory {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    description: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new TemplateCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sortOrder = source["sortOrder"];
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
	export class Template {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    name: string;
	    categoryId: number;
	    category: TemplateCategory;
	    description: string;
	    language: string;
	    code: string;
	    brokerPlatform: string;
	    strategyType: string;
	    scriptCategory: string;
	    styleTags: string;
	    emotionTags: string;
	    volumeTags: string;
	    scenarioTags: string;
	    capitalTags: string;
	    factorTags: string;
	    searchKeywords: string;
	    sourcePlatforms: string;
	    isAiGenerated: boolean;
	    aiPrompt: string;
	    aiModel: string;
	    linkedStocks: string;
	    parameters: string;
	    backtestResult: string;
	    version: number;
	    status: string;
	    // Go type: time
	    lastUsedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.name = source["name"];
	        this.categoryId = source["categoryId"];
	        this.category = this.convertValues(source["category"], TemplateCategory);
	        this.description = source["description"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.brokerPlatform = source["brokerPlatform"];
	        this.strategyType = source["strategyType"];
	        this.scriptCategory = source["scriptCategory"];
	        this.styleTags = source["styleTags"];
	        this.emotionTags = source["emotionTags"];
	        this.volumeTags = source["volumeTags"];
	        this.scenarioTags = source["scenarioTags"];
	        this.capitalTags = source["capitalTags"];
	        this.factorTags = source["factorTags"];
	        this.searchKeywords = source["searchKeywords"];
	        this.sourcePlatforms = source["sourcePlatforms"];
	        this.isAiGenerated = source["isAiGenerated"];
	        this.aiPrompt = source["aiPrompt"];
	        this.aiModel = source["aiModel"];
	        this.linkedStocks = source["linkedStocks"];
	        this.parameters = source["parameters"];
	        this.backtestResult = source["backtestResult"];
	        this.version = source["version"];
	        this.status = source["status"];
	        this.lastUsedAt = this.convertValues(source["lastUsedAt"], null);
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
	export class LinkageAIRequest {
	    summary: string;
	    templates: Template[];
	
	    static createFrom(source: any = {}) {
	        return new LinkageAIRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.templates = this.convertValues(source["templates"], Template);
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
	export class ScriptSearchRequest {
	    query: string;
	    sources: string[];
	    resultLimit: number;
	    requirePython: boolean;
	    preferPlatform: string;
	
	    static createFrom(source: any = {}) {
	        return new ScriptSearchRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.sources = source["sources"];
	        this.resultLimit = source["resultLimit"];
	        this.requirePython = source["requirePython"];
	        this.preferPlatform = source["preferPlatform"];
	    }
	}
	export class SearchLink {
	    name: string;
	    url: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.description = source["description"];
	    }
	}
	export class TagOption {
	    label: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new TagOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	    }
	}
	export class TagGroup {
	    key: string;
	    label: string;
	    description: string;
	    options: TagOption[];
	
	    static createFrom(source: any = {}) {
	        return new TagGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.options = this.convertValues(source["options"], TagOption);
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

