package resources

import (
	"github.com/chanzuckerberg/terraform-provider-snowflake/pkg/snowflake"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var ValidDatabasePrivileges = NewPrivilegeSet(
	privilegeCreateSchema,
	privilegeImportedPrivileges,
	privilegeModify,
	privilegeMonitor,
	privilegeOwnership,
	privilegeReferenceUsage,
	privilegeUsage,
)

var databaseGrantSchema = map[string]*schema.Schema{
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the database on which to grant privileges.",
		ForceNew:    true,
	},
	"privilege": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The privilege to grant on the database.",
		Default:      "USAGE",
		ValidateFunc: validation.StringInSlice(ValidDatabasePrivileges.toList(), true),
		ForceNew:     true,
	},
	"roles": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "Grants privilege to these roles.",
		ForceNew:    true,
	},
	"shares": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "Grants privilege to these shares.",
		ForceNew:    true,
	},
	"with_grant_option": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "When this is set to true, allows the recipient role to grant the privileges to other roles.",
		Default:     false,
		ForceNew:    true,
	},
}

// DatabaseGrant returns a pointer to the resource representing a database grant
func DatabaseGrant() *schema.Resource {
	return &schema.Resource{
		Create: CreateDatabaseGrant,
		Read:   ReadDatabaseGrant,
		Delete: DeleteDatabaseGrant,

		Schema: databaseGrantSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// CreateDatabaseGrant implements schema.CreateFunc
func CreateDatabaseGrant(d *schema.ResourceData, meta interface{}) error {
	dbName := d.Get("database_name").(string)
	builder := snowflake.DatabaseGrant(dbName)
	priv := d.Get("privilege").(string)
	grantOption := d.Get("with_grant_option").(bool)

	err := createGenericGrant(d, meta, builder)
	if err != nil {
		return err
	}

	grant := &grantID{
		ResourceName: dbName,
		Privilege:    priv,
		GrantOption:  grantOption,
	}
	dataIDInput, err := grant.String()
	if err != nil {
		return err
	}
	d.SetId(dataIDInput)

	return ReadDatabaseGrant(d, meta)
}

// ReadDatabaseGrant implements schema.ReadFunc
func ReadDatabaseGrant(d *schema.ResourceData, meta interface{}) error {
	grantID, err := grantIDFromString(d.Id())
	if err != nil {
		return err
	}
	err = d.Set("database_name", grantID.ResourceName)
	if err != nil {
		return err
	}
	err = d.Set("privilege", grantID.Privilege)
	if err != nil {
		return err
	}
	err = d.Set("with_grant_option", grantID.GrantOption)
	if err != nil {
		return err
	}

	// IMPORTED PRIVILEGES is not a real resource, so we can't actually verify
	// that it is still there. Just exit for now
	if grantID.Privilege == "IMPORTED PRIVILEGES" {
		return nil
	}

	builder := snowflake.DatabaseGrant(grantID.ResourceName)
	return readGenericGrant(d, meta, databaseGrantSchema, builder, false, ValidDatabasePrivileges)
}

// DeleteDatabaseGrant implements schema.DeleteFunc
func DeleteDatabaseGrant(d *schema.ResourceData, meta interface{}) error {
	dbName := d.Get("database_name").(string)
	builder := snowflake.DatabaseGrant(dbName)

	return deleteGenericGrant(d, meta, builder)
}
