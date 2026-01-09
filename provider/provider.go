package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*VaultsecProvider)(nil)
var _ provider.ProviderWithEphemeralResources = (*VaultsecProvider)(nil)

type VaultsecProvider struct{}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &VaultsecProvider{}
	}
}

type VaultsecProviderModel struct {
	Address types.String `tfsdk:"address"`
	Token   types.String `tfsdk:"token"`
}

func (p *VaultsecProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required: true,
			},
			"token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}
func (p *VaultsecProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VaultsecProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerData := &VaultsecProviderModel{
		Address: data.Address,
		Token:   data.Token,
	}

	resp.ResourceData = providerData
	resp.DataSourceData = providerData
	resp.EphemeralResourceData = providerData
}

func (p *VaultsecProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vaultsec"
}

func (p *VaultsecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVaultSecretVersionDataSource,
	}
}

func (p *VaultsecProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *VaultsecProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewVaultSecretEphemeralResource,
	}
}
