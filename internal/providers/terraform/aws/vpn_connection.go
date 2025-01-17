package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetVPNConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpn_connection",
		RFunc: NewVPNConnection,
	}
}

func NewVPNConnection(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataProcessed *decimal.Decimal

	costComponents := []*schema.CostComponent{
		{
			Name:           "VPN connection",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonVPC"),
				ProductFamily: strPtr("Cloud Connectivity"),
			},
		},
	}

	if d.Get("transit_gateway_id").String() != "" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Transit gateway attachment",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(region),
				Service:    strPtr("AmazonVPC"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Hours/")},
					{Key: "operation", Value: strPtr("TransitGatewayVPN")},
				},
			},
		})

		if u != nil && u.Get("monthly_gb_data_processed.0.value").Exists() {
			gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_processed.0.value").Float()))
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Data processed",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: gbDataProcessed,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Region:     strPtr(region),
				Service:    strPtr("AmazonVPC"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/TransitGateway-Bytes/")},
					{Key: "operation", Value: strPtr("TransitGatewayVPN")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
