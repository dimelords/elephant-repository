package docformat

import (
	context "context"
	"time"

	"github.com/google/uuid"
	"github.com/ttab/docformat/rpc/repository"
	"github.com/twitchtv/twirp"
)

type APIServer struct {
	store DocStore
}

// Interface guard
var _ repository.Documents = &APIServer{}

// Delete implements repository.Documents
func (*APIServer) Delete(context.Context, *repository.DeleteDocumentRequest) (*repository.DeleteDocumentResponse, error) {
	return nil, twirp.Unimplemented.Error("not implemented yet")
}

// Get implements repository.Documents
func (a *APIServer) Get(ctx context.Context, req *repository.GetDocumentRequest) (*repository.GetDocumentResponse, error) {
	docUUID, err := validateRequiredUUIDParam(req.Uuid, "uuid")
	if err != nil {
		return nil, err
	}

	if req.Version < 0 {
		return nil, twirp.InvalidArgumentError("version",
			"cannot be a negative number")
	}

	if req.Lock {
		return nil, twirp.Unimplemented.Error("locking is not implemented yet")
	}

	if req.Version > 0 && req.Status != "" {
		return nil, twirp.InvalidArgumentError("status",
			"status cannot be specified together with a version")
	}

	meta, err := a.store.GetDocumentMeta(ctx, docUUID)
	if IsDocStoreErrorCode(err, ErrCodeNotFound) {
		return nil, twirp.NotFoundError("the document doesn't exist")
	} else if err != nil {
		return nil, twirp.Internal.Errorf(
			"failed to load document metadata: %w", err)
	}

	var version int

	switch {
	case req.Version > 0:
		version = int(req.Version)
	case req.Status != "":
		count := len(meta.Statuses[req.Status])
		if count == 0 {
			return nil, twirp.NotFoundError(
				"no such status set for the document")
		}

		version = meta.Statuses[req.Status][count-1].Version
		if version == -1 {
			return nil, twirp.NotFoundError(
				"no such status set for the document")
		}
	default:
		version = meta.CurrentVersion
	}

	doc, err := a.store.GetDocument(ctx, docUUID, version)
	if IsDocStoreErrorCode(err, ErrCodeNotFound) {
		return nil, twirp.NotFoundError("no such version")
	} else if err != nil {
		return nil, twirp.Internal.Errorf(
			"failed to load document version: %w", err)
	}

	return &repository.GetDocumentResponse{
		Document: DocumentToRPC(doc),
		Version:  int64(version),
	}, nil
}

// GetHistory implements repository.Documents
func (a *APIServer) GetHistory(
	ctx context.Context, req *repository.GetHistoryRequest,
) (*repository.GetHistoryResponse, error) {
	docUUID, err := validateRequiredUUIDParam(req.Uuid, "uuid")
	if err != nil {
		return nil, err
	}

	if req.Before != 0 && req.Before < 2 {
		return nil, twirp.InvalidArgumentError("before",
			"cannot be non-zero and less that 2")
	}

	meta, err := a.store.GetDocumentMeta(ctx, docUUID)
	if IsDocStoreErrorCode(err, ErrCodeNotFound) {
		return nil, twirp.NotFoundError("the document doesn't exist")
	}

	start := int(req.Before) - 1
	if start == 0 {
		start = meta.Updates[len(meta.Updates)-1].Version
	}

	var res repository.GetHistoryResponse

	for i := len(meta.Updates) - 1; i >= 0; i-- {
		if meta.Updates[i].Version > start {
			continue
		}

		up := meta.Updates[i]

		res.Versions = append(res.Versions, &repository.DocumentVersion{
			Version: int64(up.Version),
			Created: up.Created.Format(time.RFC3339),
			Creator: IdentityReferenceToRPC(up.Updater),
			Meta:    UpdateMetaToRPC(up.Meta),
		})

		if len(res.Versions) == 10 {
			break
		}
	}

	return &res, nil
}

// GetMeta implements repository.Documents
func (a *APIServer) GetMeta(ctx context.Context, req *repository.GetMetaRequest) (*repository.GetMetaResponse, error) {
	docUUID, err := validateRequiredUUIDParam(req.Uuid, "uuid")
	if err != nil {
		return nil, err
	}

	meta, err := a.store.GetDocumentMeta(ctx, docUUID)
	if IsDocStoreErrorCode(err, ErrCodeNotFound) {
		return nil, twirp.NotFoundError("the document doesn't exist")
	}

	resp := repository.DocumentMeta{
		Created:        meta.Created.Format(time.RFC3339),
		Modified:       meta.Modified.Format(time.RFC3339),
		CurrentVersion: int64(meta.CurrentVersion),
	}

	for status := range meta.Statuses {
		if len(meta.Statuses[status]) == 0 {
			continue
		}

		if resp.Heads == nil {
			resp.Heads = make(map[string]*repository.Status)
		}

		statusCount := len(meta.Statuses[status])
		head := meta.Statuses[status][statusCount-1]

		s := repository.Status{
			Id:      int64(statusCount),
			Version: int64(head.Version),
			Creator: IdentityReferenceToRPC(head.Updater),
			Created: head.Created.Format(time.RFC3339),
			Meta:    UpdateMetaToRPC(head.Meta),
		}

		resp.Heads[status] = &s
	}

	for _, acl := range meta.ACL {
		resp.ALC = append(resp.ALC, &repository.ACLEntry{
			Uri:         acl.URI,
			Name:        acl.Name,
			Permissions: acl.Permissions,
		})
	}

	return &repository.GetMetaResponse{
		Meta: &resp,
	}, nil
}

func validateRequiredUUIDParam(v string, name string) (string, error) {
	if v == "" {
		return "", twirp.RequiredArgumentError(name)
	}

	u, err := uuid.Parse(v)
	if err != nil {
		return "", twirp.InvalidArgumentError(name, err.Error())
	}

	return u.String(), nil
}

// Update implements repository.Documents
func (*APIServer) Update(context.Context, *repository.UpdateRequest) (*repository.UpdateResponse, error) {
	return nil, twirp.Unimplemented.Error("not implemented yet")
}

// UpdatePermissions implements repository.Documents
func (*APIServer) UpdatePermissions(context.Context, *repository.UpdatePermissionsRequest) (*repository.UpdatePermissionsResponse, error) {
	return nil, twirp.Unimplemented.Error("not implemented yet")
}

func IdentityReferenceToRPC(ref IdentityReference) *repository.IdentityReference {
	return &repository.IdentityReference{
		Uri:  ref.URI,
		Name: ref.Name,
	}
}

func RPCToIdentityReference(ref *repository.IdentityReference) IdentityReference {
	return IdentityReference{
		URI:  ref.Uri,
		Name: ref.Name,
	}
}

func UpdateMetaToRPC(meta []UpdateMeta) []*repository.MetaValue {
	var out []*repository.MetaValue

	for i := range meta {
		out = append(out, &repository.MetaValue{
			Key:   meta[i].Key,
			Value: meta[i].Value,
		})
	}

	return out
}

func RPCToUpdateMeta(meta []*repository.MetaValue) []UpdateMeta {
	var out []UpdateMeta

	for i := range meta {
		out = append(out, UpdateMeta{
			Key:   meta[i].Key,
			Value: meta[i].Value,
		})
	}

	return out
}

func DocumentToRPC(doc *Document) *repository.Document {
	if doc == nil {
		return nil
	}

	return &repository.Document{
		Uuid:     doc.UUID,
		Type:     doc.Type,
		Uri:      doc.URI,
		Url:      doc.URL,
		Title:    doc.Title,
		Content:  BlocksToRPC(doc.Content),
		Meta:     BlocksToRPC(doc.Meta),
		Links:    BlocksToRPC(doc.Links),
		Language: doc.Language,
	}
}

func BlocksToRPC(blocks []Block) []*repository.Block {
	var res []*repository.Block

	// Not allocating up-front to avoid turning nil into [].
	if len(blocks) > 0 {
		res = make([]*repository.Block, len(blocks))
	}

	for i, b := range blocks {
		rb := repository.Block{
			Id:          b.ID,
			Uuid:        b.UUID,
			Uri:         b.URI,
			Url:         b.URL,
			Type:        b.Type,
			Title:       b.Title,
			Rel:         b.Rel,
			Role:        b.Role,
			Name:        b.Name,
			Value:       b.Value,
			ContentType: b.ContentType,
			Links:       BlocksToRPC(b.Links),
			Content:     BlocksToRPC(b.Content),
			Meta:        BlocksToRPC(b.Meta),
		}

		for k, v := range b.Data {
			if rb.Data == nil {
				rb.Data = make(map[string]string)
			}

			rb.Data[k] = v
		}

		res[i] = &rb
	}

	return res
}

func RPCToDocument(rpcDoc *repository.Document) *Document {
	if rpcDoc == nil {
		return nil
	}

	return &Document{
		UUID:     rpcDoc.Uuid,
		Type:     rpcDoc.Type,
		URI:      rpcDoc.Uri,
		URL:      rpcDoc.Url,
		Title:    rpcDoc.Title,
		Links:    RPCToBlocks(rpcDoc.Links),
		Content:  RPCToBlocks(rpcDoc.Content),
		Meta:     RPCToBlocks(rpcDoc.Meta),
		Language: rpcDoc.Language,
	}
}

func RPCToBlocks(blocks []*repository.Block) []Block {
	var res []Block

	// Not allocating up-front to avoid turning nil into [].
	if len(blocks) > 0 {
		res = make([]Block, len(blocks))
	}

	for i, rb := range blocks {
		if rb == nil {
			continue
		}

		b := Block{
			ID:          rb.Id,
			UUID:        rb.Uuid,
			URI:         rb.Uri,
			URL:         rb.Url,
			Type:        rb.Type,
			Title:       rb.Title,
			Rel:         rb.Rel,
			Role:        rb.Role,
			Name:        rb.Name,
			Value:       rb.Value,
			ContentType: rb.ContentType,
			Links:       RPCToBlocks(rb.Links),
			Content:     RPCToBlocks(rb.Content),
			Meta:        RPCToBlocks(rb.Meta),
		}

		for k, v := range rb.Data {
			if b.Data == nil {
				b.Data = make(BlockData)
			}

			b.Data[k] = v
		}

		res[i] = b
	}

	return res
}
