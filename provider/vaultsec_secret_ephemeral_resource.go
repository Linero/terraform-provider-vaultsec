package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResource = &VaultSecretEphemeralResource{}

func NewVaultSecretEphemeralResource() ephemeral.EphemeralResource {
	return &VaultSecretEphemeralResource{}
}

type VaultSecretEphemeralResource struct {
	providerData *VaultsecProviderModel
}

func (e *VaultSecretEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	e.providerData = req.ProviderData.(*VaultsecProviderModel)
}

func (e *VaultSecretEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

type VaultSecretEphemeralResourceModel struct {
	Mount           types.String `tfsdk:"mount"`
	Name            types.String `tfsdk:"name"`
	Key             types.String `tfsdk:"key"`
	PasswordLen     types.Int64  `tfsdk:"password_len"`
	OverrideSpecial types.String `tfsdk:"override_special"`
	Password        types.String `tfsdk:"password"`
	Version         types.Int64  `tfsdk:"version"`
}

func (e *VaultSecretEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"mount":            schema.StringAttribute{Required: true},
			"name":             schema.StringAttribute{Required: true},
			"key":              schema.StringAttribute{Optional: true},
			"override_special": schema.StringAttribute{Optional: true},
			"password_len": schema.Int64Attribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (e *VaultSecretEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {

	var model VaultSecretEphemeralResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	address := e.providerData.Address.ValueString()
	token := e.providerData.Token.ValueString()

	mount := model.Mount.ValueString()
	name := model.Name.ValueString()
	key := model.Key.ValueString()
	if key == "" {
		key = "password"
	}

	length := int(model.PasswordLen.ValueInt64())
	specials := model.OverrideSpecial.ValueString()

	client := &http.Client{}
	url := fmt.Sprintf("%s/v1/%s/data/%s", address, mount, name)
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

		data := result["data"].(map[string]interface{})["data"].(map[string]interface{})
		model.Password = types.StringValue(data[key].(string))
		model.Version = types.Int64Value(int64(result["data"].(map[string]interface{})["metadata"].(map[string]interface{})["version"].(float64)))
	default:
		model.Password = types.StringValue(randomPassword(length, specials))
		model.Version = types.Int64Value(1)
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &model)...)
}

func randomPassword(length int, specials ...string) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	if len(specials) > 0 {
		charset += specials[0]
	} else {
		charset += "!@#$%^&*()"
	}
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}
	return string(b)
}
