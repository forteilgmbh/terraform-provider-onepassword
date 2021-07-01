package onepassword

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceItemDocument() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceItemDocumentRead,
		Schema:      resourceItemDocument().Schema,
	}
}
