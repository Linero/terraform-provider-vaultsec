package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VaultSecretVersionDataSource{}

func NewVaultSecretVersionDataSource() datasource.DataSource {
	return &VaultSecretVersionDataSource{}
}

type VaultSecretVersionDataSource struct {
	providerData *VaultsecProviderModel
}

type VaultSecretVersionDataSourceModel struct {
	Mount   types.String `tfsdk:"mount"`
	Name    types.String `tfsdk:"name"`
	Key     types.String `tfsdk:"key"`
	Version types.Int64  `tfsdk:"version"`
}

func (d *VaultSecretVersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_version"
}

func (d *VaultSecretVersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"mount": schema.StringAttribute{Required: true},
			"name":  schema.StringAttribute{Required: true},
			"key":   schema.StringAttribute{Optional: true},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *VaultSecretVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.providerData = req.ProviderData.(*VaultsecProviderModel)
}

func (d *VaultSecretVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model VaultSecretVersionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	address := d.providerData.Address.ValueString()
	token := d.providerData.Token.ValueString()

	mount := model.Mount.ValueString()
	name := model.Name.ValueString()
	key := model.Key.ValueString()
	if key == "" {
		key = "password"
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", address, mount, name)
	reqHttp, _ := http.NewRequest("GET", url, nil)
	reqHttp.Header.Set("X-Vault-Token", token)

	httpResp, err := client.Do(reqHttp)
	if err != nil {
		resp.Diagnostics.AddError("vault request failed", err.Error())
		return
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case 200:
		body, _ := io.ReadAll(httpResp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		model.Version = types.Int64Value(int64(result["data"].(map[string]interface{})["current_version"].(float64)))
	default:
		model.Version = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
