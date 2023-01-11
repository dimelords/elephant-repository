package docformat

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Planning struct {
	XMLName     xml.Name       `xml:"planning"`
	GUID        string         `xml:"guid,attr"`
	Headline    string         `xml:"headline"`
	ItemClass   NMLQCode       `xml:"itemClass"`
	Description NMLDescription `xml:"description"`
	Properties  []ExtProperty  `xml:"planningExtProperty"`
	Links       []NMLBlock     `xml:"links>link"`
}

type NMLQCode struct {
	QCode string `xml:"qcode,attr"`
}

type NMLDescription struct {
	Role string `xml:"role,attr"`
	Text string `xml:",chardata"`
}

type ExtProperty struct {
	Type    string `xml:"type,attr"`
	Value   string `xml:"value,attr"`
	Literal string `xml:"literal,attr"`
	Why     string `xml:"why,attr"`
}

type NMLBlock struct {
	ID      string     `xml:"id,attr"`
	Title   string     `xml:"title,attr"`
	Type    string     `xml:"type,attr"`
	Rel     string     `xml:"rel,attr"`
	URI     string     `xml:"uri,attr"`
	URL     string     `xml:"url,attr"`
	UUID    string     `xml:"uuid,attr"`
	Data    NMLData    `xml:"data"`
	Links   []NMLBlock `xml:"links>link"`
	Meta    []NMLBlock `xml:"meta>object"`
	Content []NMLBlock `xml:"content>object"`
}

type NMLData struct {
	Items []NMLDataElement `xml:",any"`
}

type NMLDataElement struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func assignmentImport(
	ctx context.Context, evt OCLogEvent, opt IngestOptions,
	ccaImport converterFunc,
) (*ConvertedDoc, error) {
	var assignment Planning

	_, err := opt.Objects.GetObject(
		ctx, evt.UUID, evt.Content.Version, &assignment)
	if err != nil {
		if IsHTTPErrorWithStatus(err, http.StatusNotFound) {
			return nil, errDeletedInSource
		}

		// TODO: dirty solution to dirty data in stage, let's revisit.
		if strings.Contains(err.Error(), "XML syntax error") {
			opt.Blocklist.Add(evt.UUID, fmt.Errorf(
				"failed to load planning item: %v", err,
			))

			return nil, errIgnoreDocument
		}

		return nil, fmt.Errorf("failed to fetch original document: %w", err)
	}

	var doc Document

	doc.Language = opt.DefaultLanguage
	doc.Title = assignment.Headline
	doc.Type = "core/assignment"
	doc.UUID = assignment.GUID
	doc.URI = fmt.Sprintf("core://assignment/%s", assignment.GUID)

	var assignmentTypes []string

	switch assignment.ItemClass.QCode {
	case "plinat:newscoverage":
		// Hand over newscoverage to the standard CCA processor.
		return ccaImport(ctx, evt)
	case "ninat:composite":
		assignmentTypes = append(assignmentTypes, "video", "picture")
	default:
		assignmentTypes = append(assignmentTypes, strings.TrimPrefix(
			assignment.ItemClass.QCode, "ninat:"))
	}

	for _, t := range assignmentTypes {
		doc.Meta = append(doc.Meta, Block{
			Type:  "core/assignment-type",
			Value: t,
		})
	}

	var (
		out = ConvertedDoc{
			Status: "draft",
		}
		planningMeta = Block{
			Type: "core/assignment",
			Data: make(BlockData),
		}
	)

	for _, prop := range assignment.Properties {
		switch prop.Type {
		case "imext:status":
			out.Status = strings.TrimPrefix(prop.Value, "imext:")
		case "imext:type":
			if prop.Value != "x-im/assignment" {
				return nil, fmt.Errorf(
					"unexpected type %q for assignment",
					prop.Value,
				)
			}
		case "nrpdate:start":
			t, date, granularity, err := parseNRPDate(prop)
			if err != nil {
				return nil, err
			}

			planningMeta.Data["startDate"] = date
			planningMeta.Data["start"] = t.Format(time.RFC3339)
			planningMeta.Data["dateGranularity"] = granularity

		case "nrpdate:end":
			t, date, _, err := parseNRPDate(prop)
			if err != nil {
				return nil, err
			}

			planningMeta.Data["endDate"] = date
			planningMeta.Data["end"] = t.Format(time.RFC3339)
		}
	}

	doc.Meta = append(doc.Meta, planningMeta)

	if assignment.Description.Text != "" {
		if assignment.Description.Role != "nrpdesc:intern" {
			return nil, fmt.Errorf(
				"unexpected description role: %q",
				assignment.Description.Role)
		}

		doc.Meta = append(doc.Meta, Block{
			Type: "core/note",
			Role: "internal",
			Data: BlockData{
				"text": assignment.Description.Text,
			},
		})
	}

	for _, link := range assignment.Links {
		switch link.Rel {
		case "updater":
			out.Updater = IdentityReference{
				Name: link.Title,
				URI:  link.URI,
			}
		case "creator":
			out.Creator = IdentityReference{
				Name: link.Title,
				URI:  link.URI,
			}

			unit, ok := newsMLUnitReference(link)
			if ok {
				out.Units = append(out.Units, unit)
			}
		case "assigned-to":
			block := Block{
				Title: link.Title,
				Rel:   "assigned-to",
				Type:  "core/author",
				UUID:  link.UUID,
				Data:  make(BlockData),
			}

			for _, kv := range link.Data.Items {
				block.Data[kv.XMLName.Local] = kv.Value
			}

			doc.Links = append(doc.Links, block)
		}
	}

	out.Document = doc

	return &out, nil
}

func parseNRPDate(prop ExtProperty) (time.Time, string, string, error) {
	granularity := strings.TrimPrefix(prop.Why, "nrpwhy:")

	return parseGranularTime(prop.Value, prop.Type, granularity)
}

func parseGranularTime(
	value, name, granularity string,
) (time.Time, string, string, error) {
	layout := time.RFC3339

	// Heuristic for finding undeclared date granularity values.
	if granularity == "" && len(value) == 10 {
		granularity = "date"
	}

	if granularity == "date" && len(value) == 10 {
		layout = "2006-01-02"
	}

	if granularity == "" {
		granularity = "datetime"
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, "", "", fmt.Errorf(
			"invalid %q timestamp: %w",
			name, err)
	}

	return t.UTC(), t.Format("2006-01-02"), granularity, nil
}

func newsMLUnitReference(link NMLBlock) (IdentityReference, bool) {
	for _, l := range link.Links {
		if l.Type != "x-imid/organisation" ||
			l.Rel != "affiliation" {
			continue
		}

		for _, u := range l.Links {
			if u.Type != "x-imid/unit" ||
				u.Rel != "affiliation" {
				continue
			}

			return IdentityReference{
				Name: u.Title,
				URI: strings.Replace(u.URI,
					"imid://unit/", "core://unit/", 1),
			}, true
		}
	}

	return IdentityReference{}, false
}
