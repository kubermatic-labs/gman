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
		SkuId:     "1010020020", // G Suite Enterprise
		Name:      "GSuiteEnterprise",
	},
	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-Unlimited", // G Suite Business
		Name:      "GSuiteBusiness",
	},
	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-For-Business", // G Suite Basic
		Name:      "GSuiteBasic",
	},
	{
		ProductId: "101006",
		SkuId:     "1010060001", // G Suite Essentials
		Name:      "GSuiteEssentials",
	},
	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-Lite", // G Suite Lite
		Name:      "GSuiteLite",
	},
	{
		ProductId: "Google-Apps",
		SkuId:     "Google-Apps-For-Postini", // Google Apps Message Security
		Name:      "GoogleAppsMessageSecurity",
	},
	{
		ProductId: "101031",     // G Suite Enterprise for Education
		SkuId:     "1010310002", // G Suite Enterprise for Education
		Name:      "GSuiteEducation",
	},
	{
		ProductId: "101031",     // G Suite Enterprise for Education
		SkuId:     "1010310003", // G Suite Enterprise for Education (Student)
		Name:      "GSuiteEducationStudent",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-20GB",
		Name:      "GoogleDrive20GB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-50GB",
		Name:      "GoogleDrive50GB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-200GB",
		Name:      "GoogleDrive200GB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-400GB",
		Name:      "GoogleDrive400GB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-1TB",
		Name:      "GoogleDrive1TB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-2TB",
		Name:      "GoogleDrive2TB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-4TB",
		Name:      "GoogleDrive4TB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-8TB",
		Name:      "GoogleDrive8TB",
	},
	{
		ProductId: "Google-Drive-storage",
		SkuId:     "Google-Drive-storage-16TB",
		Name:      "GoogleDrive16TB",
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
		ProductId: "101005", // Cloud Identity Premium
		SkuId:     "1010050001",
		Name:      "CloudIdentityPremium",
	},
	{
		ProductId: "101033", // Google Voice
		SkuId:     "1010330003",
		Name:      "GoogleVoiceStarter",
	},
	{
		ProductId: "101033", // Google Voice
		SkuId:     "1010330004",
		Name:      "GoogleVoiceStandard",
	},
	{
		ProductId: "101033", // Google Voice
		SkuId:     "1010330002",
		Name:      "GoogleVoicePremier",
	},
}
