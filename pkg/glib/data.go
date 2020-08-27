// Package glib contains methods for interactions with GSuite API
// This file contains static-data for the methods
package glib

type License struct {
	productId string
	skuId     string
	name      string // used in yaml
}

var googleLicenses = []License{
	{
		productId: "Google-Apps",
		skuId:     "1010020020", // G Suite Enterprise
		name:      "GSuiteEnterprise",
	},
	{
		productId: "Google-Apps",
		skuId:     "Google-Apps-Unlimited", // G Suite Business
		name:      "GSuiteBusiness",
	},
	{
		productId: "Google-Apps",
		skuId:     "Google-Apps-For-Business", // G Suite Basic
		name:      "GSuiteBasic",
	},
	{
		productId: "101006",
		skuId:     "1010060001", // G Suite Essentials
		name:      "GSuiteEssentials",
	},
	{
		productId: "Google-Apps",
		skuId:     "Google-Apps-Lite", // G Suite Lite
		name:      "GSuiteLite",
	},
	{
		productId: "Google-Apps",
		skuId:     "Google-Apps-For-Postini", // Google Apps Message Security
		name:      "GoogleAppsMessageSecurity",
	},
	{
		productId: "101031",     // G Suite Enterprise for Education
		skuId:     "1010310002", // G Suite Enterprise for Education
		name:      "GSuiteEducation",
	},
	{
		productId: "101031",     // G Suite Enterprise for Education
		skuId:     "1010310003", // G Suite Enterprise for Education (Student)
		name:      "GSuiteEducationStudent",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-20GB",
		name:      "GoogleDrive20GB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-50GB",
		name:      "GoogleDrive50GB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-200GB",
		name:      "GoogleDrive200GB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-400GB",
		name:      "GoogleDrive400GB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-1TB",
		name:      "GoogleDrive1TB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-2TB",
		name:      "GoogleDrive2TB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-4TB",
		name:      "GoogleDrive4TB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-8TB",
		name:      "GoogleDrive8TB",
	},
	{
		productId: "Google-Drive-storage",
		skuId:     "Google-Drive-storage-16TB",
		name:      "GoogleDrive16TB",
	},
	{
		productId: "Google-Vault",
		skuId:     "Google-Vault",
		name:      "GoogleVault",
	},
	{
		productId: "Google-Vault",
		skuId:     "Google-Vault-Former-Employee",
		name:      "GoogleVaultFormerEmployee",
	},
	{
		productId: "101001", // Cloud Identity
		skuId:     "1010010001",
		name:      "CloudIdentity",
	},
	{
		productId: "101005", // Cloud Identity Premium
		skuId:     "1010050001",
		name:      "CloudIdentityPremium",
	},
	{
		productId: "101033", // Google Voice
		skuId:     "1010330003",
		name:      "GoogleVoiceStarter",
	},
	{
		productId: "101033", // Google Voice
		skuId:     "1010330004",
		name:      "GoogleVoiceStandard",
	},
	{
		productId: "101033", // Google Voice
		skuId:     "1010330002",
		name:      "GoogleVoicePremier",
	},
}
