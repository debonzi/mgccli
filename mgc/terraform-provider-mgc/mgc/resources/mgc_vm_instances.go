package resources

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	bws "github.com/geffersonFerraz/brazilian-words-sorter"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mgcSdk "magalu.cloud/lib"
	sdkVmImages "magalu.cloud/lib/products/virtual_machine/images"
	sdkVmInstances "magalu.cloud/lib/products/virtual_machine/instances"
	sdkVmMachineTypes "magalu.cloud/lib/products/virtual_machine/machine_types"
	"magalu.cloud/terraform-provider-mgc/mgc/client"
	tfutil "magalu.cloud/terraform-provider-mgc/mgc/tfutil"
)

var (
	_ resource.Resource              = &vmInstances{}
	_ resource.ResourceWithConfigure = &vmInstances{}
)

func NewVirtualMachineInstancesResource() resource.Resource {
	return &vmInstances{}
}

type vmInstances struct {
	sdkClient      *mgcSdk.Client
	vmInstances    sdkVmInstances.Service
	vmImages       sdkVmImages.Service
	vmMachineTypes sdkVmMachineTypes.Service
}

func (r *vmInstances) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine_instances"
}

func (r *vmInstances) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	var err error
	var errDetail error
	r.sdkClient, err, errDetail = client.NewSDKClient(req)
	if err != nil {
		resp.Diagnostics.AddError(
			err.Error(),
			errDetail.Error(),
		)
		return
	}

	r.vmInstances = sdkVmInstances.NewService(ctx, r.sdkClient)
	r.vmImages = sdkVmImages.NewService(ctx, r.sdkClient)
	r.vmMachineTypes = sdkVmMachineTypes.NewService(ctx, r.sdkClient)
}

type vmInstancesResourceModel struct {
	ID           types.String                `tfsdk:"id"`
	Name         types.String                `tfsdk:"name"`
	NameIsPrefix types.Bool                  `tfsdk:"name_is_prefix"`
	FinalName    types.String                `tfsdk:"final_name"`
	UpdatedAt    types.String                `tfsdk:"updated_at"`
	CreatedAt    types.String                `tfsdk:"created_at"`
	SshKeyName   types.String                `tfsdk:"ssh_key_name"`
	State        types.String                `tfsdk:"state"`
	Status       types.String                `tfsdk:"status"`
	Network      networkVmInstancesModel     `tfsdk:"network"`
	MachineType  vmInstancesMachineTypeModel `tfsdk:"machine_type"`
	Image        tfutil.GenericIDNameModel   `tfsdk:"image"`
	UserData     types.String                `tfsdk:"user_data"`
}

type networkVmInstancesModel struct {
	IPV6              types.String                          `tfsdk:"ipv6"`
	PrivateAddress    types.String                          `tfsdk:"private_address"`
	PublicIpAddress   types.String                          `tfsdk:"public_address"`
	DeletePublicIP    types.Bool                            `tfsdk:"delete_public_ip"`
	AssociatePublicIP types.Bool                            `tfsdk:"associate_public_ip"`
	VPC               *tfutil.GenericIDNameModel            `tfsdk:"vpc"`
	Interface         *vmInstancesNetworkSecurityGroupModel `tfsdk:"interface"`
}

type vmInstancesNetworkSecurityGroupModel struct {
	SecurityGroups []tfutil.GenericIDModel `tfsdk:"security_groups"`
}

type vmInstancesMachineTypeModel struct {
	ID    types.String `tfsdk:"id"`
	Disk  types.Number `tfsdk:"disk"`
	Name  types.String `tfsdk:"name"`
	RAM   types.Number `tfsdk:"ram"`
	VCPUs types.Number `tfsdk:"vcpus"`
}

func (r *vmInstances) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	description := "Operations with instances, including create, delete, start, stop, reboot and other actions."
	resp.Schema = schema.Schema{
		Description:         description,
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the virtual machine instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
			},
			"name_is_prefix": schema.BoolAttribute{
				Description: "Indicates whether the provided name is a prefix or the exact name of the virtual machine instance.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"name": schema.StringAttribute{
				Description: "The name of the virtual machine instance.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`),
						"The name must contain only lowercase letters, numbers, underlines and hyphens. Hyphens and underlines cannot be located at the edges either.",
					),
				},
				Required: true,
			},
			"final_name": schema.StringAttribute{
				Description: "The final name of the virtual machine instance after applying any naming conventions or modifications.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the virtual machine instance was last updated.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the virtual machine instance was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
			},
			"ssh_key_name": schema.StringAttribute{
				Description: "The name of the SSH key associated with the virtual machine instance. If the image is Windows, this field is not used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"state": schema.StringAttribute{
				Description: "The current state of the virtual machine instance.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the virtual machine instance.",
				Computed:    true,
			},
			"user_data": schema.StringAttribute{
				Description: "Used to perform automated configuration tasks. Must be sent in base64 format. (between 1 and 65000 characters)",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 65000),
					stringvalidator.RegexMatches(regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`),
						"Must be a valid base64 string."),
				},
			},
			"image": schema.SingleNestedAttribute{
				Description: "The image used to create the virtual machine instance.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "The unique identifier of the image.",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "The name of the image.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Required: true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
			"machine_type": schema.SingleNestedAttribute{
				Description: "The machine type of the virtual machine instance.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "The unique identifier of the machine type.",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "The name of the machine type.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Required: true,
					},
					"disk": schema.NumberAttribute{
						Description: "The disk size of the machine type.",
						Computed:    true,
					},
					"ram": schema.NumberAttribute{
						Description: "The RAM size of the machine type.",
						Computed:    true,
					},
					"vcpus": schema.NumberAttribute{
						Description: "The number of virtual CPUs of the machine type.",
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
			"network": schema.SingleNestedAttribute{
				Description: "The network configuration of the virtual machine instance.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"delete_public_ip": schema.BoolAttribute{
						Description: "Indicates whether to delete the public IP address associated with the virtual machine instance.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						Default:  booldefault.StaticBool(true),
						Optional: true,
						Computed: true,
					},
					"associate_public_ip": schema.BoolAttribute{
						Description: "Indicates whether to associate a public IP address with the virtual machine instance.",
						Required:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"ipv6": schema.StringAttribute{
						Description: "The IPv6 address of the virtual machine instance.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Computed: true,
					},
					"private_address": schema.StringAttribute{
						Description: "The private IP address of the virtual machine instance.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Computed: true,
					},
					"public_address": schema.StringAttribute{
						Description: "The public IP address of the virtual machine instance.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Computed: true,
					},
					"vpc": schema.SingleNestedAttribute{
						Description: "The VPC (Virtual Private Cloud) associated with the virtual machine instance.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Description: "The unique identifier of the VPC.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
								Optional: true,
								Computed: true,
							},
							"name": schema.StringAttribute{
								Description: "The name of the VPC.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
								Optional: true,
								Computed: true,
							},
						},
					},
					"interface": schema.SingleNestedAttribute{
						Description: "The network interface configuration of the virtual machine instance.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"security_groups": schema.ListNestedAttribute{
								Description: "The security groups associated with the network interface.",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "The unique identifier of the security group.",
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *vmInstances) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data := vmInstancesResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	getResult, err := r.getVmStatus(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading VM",
			"Could not read VM ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(getResult.Id)
	data = r.setValuesFromServer(data, getResult)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmInstances) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := vmInstancesResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := plan
	state.FinalName = types.StringValue(state.Name.ValueString())

	if state.NameIsPrefix.ValueBool() {
		bwords := bws.BrazilianWords(3, "-")
		state.FinalName = types.StringValue(state.Name.ValueString() + "-" + bwords.Sort())
	}

	if state.Network.DeletePublicIP.IsNull() {
		state.Network.DeletePublicIP = types.BoolValue(true)
	}

	if state.Network.AssociatePublicIP.IsNull() {
		state.Network.AssociatePublicIP = types.BoolValue(false)
	}

	createParams := sdkVmInstances.CreateParameters{
		Name:       state.FinalName.ValueString(),
		SshKeyName: state.SshKeyName.ValueStringPointer(),
		Image: sdkVmInstances.CreateParametersImage{
			Name: state.Image.Name.ValueStringPointer(),
		},
		MachineType: sdkVmInstances.CreateParametersMachineType{
			Name: state.MachineType.Name.ValueStringPointer(),
		},
		Network: &sdkVmInstances.CreateParametersNetwork{},
	}

	if state.Network.VPC != nil && state.Network.VPC.ID.ValueString() != "" {
		createParams.Network.Vpc = &sdkVmInstances.CreateParametersNetworkVpc{
			Id: state.Network.VPC.ID.ValueString(),
		}
	}

	if state.Network.Interface != nil && len(state.Network.Interface.SecurityGroups) > 0 {
		network := sdkVmInstances.CreateParametersNetwork{}
		network.Interface = &sdkVmInstances.CreateParametersNetworkInterface{}
		network.Interface.SecurityGroups = &sdkVmInstances.CreateParametersNetworkInterfaceSecurityGroups{}

		items := []sdkVmInstances.CreateParametersNetworkInterfaceSecurityGroupsItem{}

		for _, sg := range state.Network.Interface.SecurityGroups {
			items = append(items, sdkVmInstances.CreateParametersNetworkInterfaceSecurityGroupsItem{
				Id: sg.ID.ValueString(),
			})
		}
		vmInstancesNetworkInterfaceSecurityGroups := sdkVmInstances.CreateParametersNetworkInterfaceSecurityGroups(items)
		createParams.Network.Interface = &sdkVmInstances.CreateParametersNetworkInterface{}
		createParams.Network.Interface.SecurityGroups = &vmInstancesNetworkInterfaceSecurityGroups
	}

	createParams.Network.AssociatePublicIp = state.Network.AssociatePublicIP.ValueBoolPointer()
	createParams.UserData = state.UserData.ValueStringPointer()

	result, err := r.vmInstances.CreateContext(ctx, createParams, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.CreateConfigs{}))
	if err != nil {
		resp.Diagnostics.AddError("Error creating vm", err.Error())
		return
	}

	getResult, err := r.getVmStatus(ctx, result.Id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading VM", err.Error())
		return
	}
	state.ID = types.StringValue(result.Id)

	state = r.setValuesFromServer(state, getResult)

	state.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	state.UpdatedAt = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *vmInstances) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	data := vmInstancesResourceModel{}
	currState := &vmInstancesResourceModel{}
	req.State.Get(ctx, currState)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if !currState.Name.Equal(data.Name) {
		data.FinalName = types.StringValue(data.Name.ValueString())

		if data.NameIsPrefix.ValueBool() {
			bwords := bws.BrazilianWords(3, "-")
			data.FinalName = types.StringValue(data.Name.ValueString() + "-" + bwords.Sort())
		}
		err := r.vmInstances.RenameContext(ctx, sdkVmInstances.RenameParameters{
			Id:   data.ID.ValueString(),
			Name: data.FinalName.ValueString(),
		}, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.RenameConfigs{}))

		if err != nil {
			resp.Diagnostics.AddError(
				"Error to rename vm",
				"Could not rename the vm instance, unexpected error: "+err.Error(),
			)
			return
		}
	}

	machineType, err := r.getMachineTypeID(ctx, data.MachineType.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error to retype vm",
			"Could not found machine-type or load machine-type list, unexpected error: "+err.Error(),
		)
		return
	}

	if !currState.MachineType.ID.Equal(machineType.ID) {
		err = r.vmInstances.RetypeContext(ctx, sdkVmInstances.RetypeParameters{
			Id: data.ID.ValueString(),
			MachineType: sdkVmInstances.RetypeParametersMachineType{
				Id: machineType.ID.ValueString(),
			},
		}, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.RetypeConfigs{}))

		if err != nil {
			resp.Diagnostics.AddError(
				"Error on Update VM",
				"Could not update VM machine-type "+data.ID.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	data.UpdatedAt = types.StringValue(time.Now().Format(time.RFC850))
	getResult, err := r.getVmStatus(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading VM",
			"Error when get new vm status "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	data = r.setValuesFromServer(data, getResult)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmInstances) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vmInstancesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.vmInstances.DeleteContext(ctx,
		sdkVmInstances.DeleteParameters{
			DeletePublicIp: data.Network.DeletePublicIP.ValueBoolPointer(),
			Id:             data.ID.ValueString(),
		},
		tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.DeleteConfigs{}))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VM",
			"Could not delete VM ID "+data.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	r.checkVmIsNotFound(ctx, data.ID.ValueString())
}

func (r *vmInstances) setValuesFromServer(data vmInstancesResourceModel, server *sdkVmInstances.GetResult) vmInstancesResourceModel {
	data.ID = types.StringValue(server.Id)
	data.FinalName = types.StringValue(*server.Name)
	data.State = types.StringValue(server.State)
	data.Status = types.StringValue(server.Status)
	data.MachineType.ID = types.StringValue(server.MachineType.Id)
	data.MachineType.Name = types.StringValue(*server.MachineType.Name)
	data.MachineType.Disk = types.NumberValue(new(big.Float).SetInt64(int64(*server.MachineType.Disk)))
	data.MachineType.RAM = types.NumberValue(new(big.Float).SetInt64(int64(*server.MachineType.Ram)))
	data.MachineType.VCPUs = types.NumberValue(new(big.Float).SetInt64(int64(*server.MachineType.Vcpus)))

	data.Image.Name = types.StringValue(*server.Image.Name)
	data.Image.ID = types.StringValue(server.Image.Id)

	if vpc := data.Network.VPC; vpc != nil {
		vpc.ID = types.StringValue(server.Network.Vpc.Id)
		vpc.Name = types.StringValue(server.Network.Vpc.Name)
	}

	data.Network.IPV6 = types.StringValue("")
	data.Network.PrivateAddress = types.StringValue("")
	data.Network.PublicIpAddress = types.StringValue("")

	if server.Network.Ports != nil && len(*server.Network.Ports) > 0 {
		ports := (*server.Network.Ports)[0]

		data.Network.PrivateAddress = types.StringValue(ports.IpAddresses.PrivateIpAddress)

		if ports.IpAddresses.IpV6address != nil {
			data.Network.IPV6 = types.StringValue(*ports.IpAddresses.IpV6address)
		}

		if ports.IpAddresses.PublicIpAddress != nil {
			data.Network.PublicIpAddress = types.StringValue(*ports.IpAddresses.PublicIpAddress)
		}

	}
	data.UserData = types.StringPointerValue(server.UserData)
	return data
}

func (r *vmInstances) getMachineTypeID(ctx context.Context, name string) (*vmInstancesMachineTypeModel, error) {
	machineType := vmInstancesMachineTypeModel{}
	machineTypeList, err := r.vmMachineTypes.ListContext(ctx, sdkVmMachineTypes.ListParameters{}, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmMachineTypes.ListConfigs{}))
	if err != nil {
		return nil, fmt.Errorf("could not load machine-type list, unexpected error: %w", err)
	}

	for _, x := range machineTypeList.MachineTypes {
		if x.Name == name {
			machineType.Disk = types.NumberValue(new(big.Float).SetInt64(int64(x.Disk)))
			machineType.ID = types.StringValue(x.Id)
			machineType.Name = types.StringValue(x.Name)
			machineType.RAM = types.NumberValue(new(big.Float).SetInt64(int64(x.Ram)))
			machineType.VCPUs = types.NumberValue(new(big.Float).SetInt64(int64(x.Vcpus)))
			break
		}
	}

	if machineType.ID.ValueString() == "" {
		return nil, fmt.Errorf("could not found machine-type ID with name: %s", name)
	}
	return &machineType, nil
}

func (r *vmInstances) getVmStatus(ctx context.Context, id string) (*sdkVmInstances.GetResult, error) {
	getResult := &sdkVmInstances.GetResult{}
	expand := &sdkVmInstances.GetParametersExpand{"network", "machine-types", "image"}

	duration := 5 * time.Minute
	startTime := time.Now()
	getParam := sdkVmInstances.GetParameters{Id: id, Expand: expand}
	var err error
	for {
		elapsed := time.Since(startTime)
		remaining := duration - elapsed
		if remaining <= 0 {
			if getResult.Status != "" {
				return getResult, nil
			}
			return getResult, fmt.Errorf("timeout to read VM ID")
		}

		*getResult, err = r.vmInstances.GetContext(ctx, getParam, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.GetConfigs{}))
		if err != nil {
			return getResult, err
		}

		if getResult.Status == "completed" {
			return getResult, nil
		}
		time.Sleep(3 * time.Second)
	}
}

func (r *vmInstances) checkVmIsNotFound(ctx context.Context, id string) {
	getResult := &sdkVmInstances.GetResult{}

	duration := 5 * time.Minute
	startTime := time.Now()
	getParam := sdkVmInstances.GetParameters{Id: id}
	var err error
	for {
		elapsed := time.Since(startTime)
		remaining := duration - elapsed
		if remaining <= 0 {
			if getResult.Status != "" {
				return
			}
			return
		}

		*getResult, err = r.vmInstances.GetContext(ctx, getParam, tfutil.GetConfigsFromTags(r.sdkClient.Sdk().Config().Get, sdkVmInstances.GetConfigs{}))
		if err != nil {
			return
		}

		time.Sleep(3 * time.Second)
	}
}
