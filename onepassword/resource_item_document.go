package onepassword

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceItemDocument() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceItemDocumentRead,
		CreateContext: resourceItemDocumentCreate,
		DeleteContext: resourceItemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if err := resourceItemDocumentRead(ctx, d, meta); err.HasError() {
					return []*schema.ResourceData{d}, errors.New(err[0].Summary)
				}
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vault": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"file_path": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"filename", "content", "content_base64"},
			},
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"file_path", "content_base64"},
				RequiredWith:  []string{"filename"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"file_path", "content"},
				RequiredWith:  []string{"filename"},
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"file_path"},
			},
			"archived": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceItemDocumentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	m := meta.(*Meta)
	vaultID := d.Get("vault").(string)
	v, err := m.onePassClient.ReadItem(getID(d), vaultID)
	if err != nil {
		return diag.FromErr(err)
	}
	if v == nil {
		log.Printf("[INFO] Item %s not found in %s vault", getID(d), vaultID)
		d.SetId("")
		return nil
	}

	if v.Template != Category2Template(DocumentCategory) {
		return diag.FromErr(errors.New("item is not from " + string(DocumentCategory)))
	}

	d.SetId(v.UUID)
	if err := d.Set("name", v.Overview.Title); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", v.Overview.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vault", v.Vault); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filename", v.Details.DocumentAttributes.FileName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("archived", v.Trashed == IsTrashed); err != nil {
		diag.FromErr(err)
	}

	content, err := m.onePassClient.ReadDocument(v.UUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content", string(content)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content_base64", b64.StdEncoding.EncodeToString(content)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceItemDocumentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var filename string
	var fileContent []byte

	if path, ok := d.GetOk("file_path"); ok {
		filename = filepath.Base(path.(string))
		if dat, err := ioutil.ReadFile(path.(string)); err == nil {
			fileContent = dat
		} else {
			return diag.FromErr(err)
		}
	} else if contentB64, ok := d.GetOk("content_base64"); ok {
		filename = d.Get("filename").(string)
		if dec, err := b64.StdEncoding.DecodeString(contentB64.(string)); err == nil {
			fileContent = dec
		} else {
			return diag.FromErr(err)
		}
	} else {
		filename = d.Get("filename").(string)
		fileContent = []byte(d.Get("content").(string))
	}

	item := &Item{
		Vault:    d.Get("vault").(string),
		Template: Category2Template(DocumentCategory),
		Overview: Overview{
			Title: d.Get("name").(string),
			Tags:  ParseTags(d),
		},
		Details: Details{
			DocumentAttributes: DocumentAttributes{
				FileName: filename,
			},
		},
	}
	m := meta.(*Meta)
	err := m.onePassClient.CreateDocument(item, fileContent)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(item.UUID)
	return resourceItemDocumentRead(ctx, d, meta)
}
