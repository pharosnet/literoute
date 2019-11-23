package literoute

import (
	"encoding/xml"
	"io"
)

type xmlMapEntry struct {
	XMLName xml.Name
	Value   interface{} `xml:",chardata"`
}

func XMLMap(elementName string, v Map) xml.Marshaler {
	return xmlMap{
		entries:     v,
		elementName: elementName,
	}
}

type xmlMap struct {
	entries     Map
	elementName string
}

func (m xmlMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m.entries) == 0 {
		return nil
	}

	start.Name = xml.Name{Local: m.elementName}
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m.entries {
		_ = e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
	}

	return e.EncodeToken(start.End())
}

func WriteXML(writer io.Writer, v interface{}, options XML) (int, error) {
	if prefix := options.Prefix; prefix != "" {
		_, _ = writer.Write([]byte(prefix))
	}

	if indent := options.Indent; indent != "" {
		result, err := xml.MarshalIndent(v, "", indent)
		if err != nil {
			return 0, err
		}
		result = append(result, newLineB...)
		return writer.Write(result)
	}

	result, err := xml.Marshal(v)
	if err != nil {
		return 0, err
	}
	return writer.Write(result)
}
