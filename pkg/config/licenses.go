package config

type License struct {
	ProductId string
	SkuId     string
	Name      string // used in yaml
}

// list of available GSuite Licenses
var AllLicenses = []License{
	{
		ProductId: "Google-Apps",
		SkuId:     "1010020027",
		Name:      "GoogleWorkspaceBusinessStarter",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010020028",
		Name:      "GoogleWorkspaceBusinessStandard",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010020025",
		Name:      "GoogleWorkspaceBusinessPlus",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010060003",
		Name:      "GoogleWorkspaceEnterpriseEssentials",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010020026",
		Name:      "GoogleWorkspaceEnterpriseStandard",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010020020",
		Name:      "GoogleWorkspaceEnterprisePlus",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "1010060001",
		Name:      "GoogleWorkspaceEssentials",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-Unlimited",
		Name:      "GSuiteBusiness",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-For-Business",
		Name:      "GSuiteBasic",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-Lite",
		Name:      "GSuiteLite",
	},

	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-For-Postini",
		Name:      "GoogleAppsMessageSecurity",
	},

	{
		ProductId: "101031",
		SkuId:     "1010310002",
		Name:      "GSuiteEnterpriseForEducation",
	},

	{
		ProductId: "101031",
		SkuId:     "1010310003",
		Name:      "GSuiteEnterpriseForEducationStudent",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-20GB",
		Name:      "GoogleDriveStorage20GB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-50GB",
		Name:      "GoogleDriveStorage50GB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-200GB",
		Name:      "GoogleDriveStorage200GB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-400GB",
		Name:      "GoogleDriveStorage400GB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-1TB",
		Name:      "GoogleDriveStorage1TB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-2TB",
		Name:      "GoogleDriveStorage2TB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-4TB",
		Name:      "GoogleDriveStorage4TB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-8TB",
		Name:      "GoogleDriveStorage8TB",
	},

	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-16TB",
		Name:      "GoogleDriveStorage16TB",
	},

	{
		ProductId: "Google-Vault",
		SkuId:     "Google-Vault",
		Name:      "GoogleVault",
	},

	{
		ProductId: "Google-Vault",
		SkuId:     "Google-Vault-Former-Employee",
		Name:      "GoogleVaultFormerEmployee",
	},

	{
		ProductId: "101001",
		SkuId:     "1010010001",
		Name:      "CloudIdentity",
	},

	{
		ProductId: "101005",
		SkuId:     "1010050001",
		Name:      "CloudIdentityPremium",
	},

	{
		ProductId: "101033",
		SkuId:     "1010330003",
		Name:      "GoogleVoiceStarter",
	},

	{
		ProductId: "101033",
		SkuId:     "1010330004",
		Name:      "GoogleVoiceStandard",
	},

	{
		ProductId: "101033",
		SkuId:     "1010330002",
		Name:      "GoogleVoicePremier",
	},
}
