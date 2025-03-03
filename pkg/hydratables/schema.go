package hydratables

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"ecksbee.com/telefacts/pkg/attr"
	"ecksbee.com/telefacts/pkg/serializables"
)

const COMMON_SCHEMA = `http://www.xbrl.org/2003/xbrl-instance-2003-12-31.XSD`

type Concept struct {
	XMLName           xml.Name
	ID                string
	Type              xml.Name
	SubstitutionGroup xml.Name
	Nillable          bool
	PeriodType        string
	Balance           string
	Abstract          bool
	TypedDomainHref   string
}

type Schema struct {
	FileName string
	Annotation
	Element []Concept
}

type Annotation struct {
	Appinfo
}

type Appinfo struct {
	RoleTypes []RoleType
}

type RoleType struct {
	RoleURI    string
	ID         string
	Definition string
	UsedOn     []struct {
		CharData string
	}
}

func HydrateSchema(file *serializables.SchemaFile, fileName string) (*Schema, error) {
	if len(fileName) <= 0 {
		return nil, fmt.Errorf("empty file name")
	}
	if file == nil {
		return nil, fmt.Errorf("empty file")
	}
	ret := Schema{}
	ret.FileName = fileName
	ret.Annotation = hydrateAnnotation(file)
	ret.Element = hydrateConcepts(file, fileName)
	return &ret, nil
}

func hydrateTargetNamespace(file *serializables.SchemaFile) string {
	attr := attr.FindAttr(file.XMLAttrs, "targetNamespace")
	if attr == nil {
		return ""
	}
	return attr.Value
}

func hydrateElementFormDefault(file *serializables.SchemaFile) string {
	attr := attr.FindAttr(file.XMLAttrs, "elementFormDefault")
	if attr == nil {
		return ""
	}
	return attr.Value
}

func hydrateAttributeFormDefault(file *serializables.SchemaFile) string {
	attr := attr.FindAttr(file.XMLAttrs, "attributeFormDefault")
	if attr == nil {
		return ""
	}
	return attr.Value
}

func hydrateAnnotation(file *serializables.SchemaFile) Annotation {
	ret := Annotation{}
	if len(file.Annotation) <= 0 {
		return ret
	}
	if file.Annotation[0].XMLName.Space != attr.XSD {
		return ret
	}
	if len(file.Annotation[0].Appinfo) <= 0 {
		return ret
	}
	if file.Annotation[0].Appinfo[0].XMLName.Space != attr.XSD {
		return ret
	}
	roleTypes := make([]RoleType, 0, len(file.Annotation[0].Appinfo[0].RoleType))
	for _, roleType := range file.Annotation[0].Appinfo[0].RoleType {
		if roleType.XMLName.Space != attr.LINK {
			continue
		}
		newRoleType := RoleType{}
		idAttr := attr.FindAttr(roleType.XMLAttrs, "id")
		if idAttr == nil || idAttr.Value == "" {
			continue
		}
		newRoleType.ID = idAttr.Value
		roleuriAttr := attr.FindAttr(roleType.XMLAttrs, "roleURI")
		if roleuriAttr == nil || roleuriAttr.Value == "" {
			continue
		}
		newRoleType.RoleURI = roleuriAttr.Value
		if len(roleType.Definition) > 0 {
			newRoleType.Definition = roleType.Definition[0].CharData
		}
		if len(roleType.UsedOn) > 0 {
			newRoleType.UsedOn = make([]struct{ CharData string }, 0, len(roleType.UsedOn))
			for _, usedOn := range roleType.UsedOn {
				if usedOn.XMLName.Space != attr.LINK {
					continue
				}
				newRoleType.UsedOn = append(newRoleType.UsedOn, struct{ CharData string }{
					CharData: usedOn.CharData,
				})
			}
		}
		roleTypes = append(roleTypes, newRoleType)
	}
	return Annotation{
		Appinfo: Appinfo{
			RoleTypes: roleTypes,
		},
	}
}

func hydrateConcepts(file *serializables.SchemaFile, fileName string) []Concept {
	ret := make([]Concept, 0, len(file.Element))
	tlAttrs := file.XMLAttrs
	targetNS := hydrateTargetNamespace(file)
	for _, element := range file.Element {
		if element.XMLName.Space != attr.XSD {
			continue
		}
		idAttr := attr.FindAttr(element.XMLAttrs, "id")
		if idAttr == nil || idAttr.Value == "" {
			continue
		}
		nameAttr := attr.FindAttr(element.XMLAttrs, "name")
		if nameAttr == nil || nameAttr.Value == "" {
			continue
		}
		typeAttr := attr.FindAttr(element.XMLAttrs, "type")
		if typeAttr == nil || typeAttr.Value == "" {
			continue
		}
		substitutionGroup := xml.Name{}
		substitutionGroupAttr := attr.FindAttr(element.XMLAttrs, "substitutionGroup")
		if substitutionGroupAttr != nil && substitutionGroupAttr.Value != "" {
			substitutionGroup = attr.Xmlns(tlAttrs, substitutionGroupAttr.Value)
		}
		isAbstract := false
		abstractAttr := attr.FindAttr(element.XMLAttrs, "abstract")
		if abstractAttr != nil {
			v, err := strconv.ParseBool(abstractAttr.Value)
			if err == nil {
				isAbstract = v
			}
		}
		balance := ""
		balanceAttr := attr.FindAttr(element.XMLAttrs, "balance")
		if balanceAttr != nil && balanceAttr.Name.Space == attr.XBRLI {
			balance = balanceAttr.Value
		}
		isNillable := false
		nillableAttr := attr.FindAttr(element.XMLAttrs, "nillable")
		if nillableAttr != nil {
			v, err := strconv.ParseBool(nillableAttr.Value)
			if err == nil {
				isNillable = v
			}
		}
		periodType := ""
		periodTypeAttr := attr.FindAttr(element.XMLAttrs, "periodType")
		if periodTypeAttr != nil && periodTypeAttr.Name.Space == attr.XBRLI {
			periodType = periodTypeAttr.Value
		}
		typedDomainHref := ""
		typedDomainRefAttr := attr.FindAttr(element.XMLAttrs, "typedDomainRef")
		if typedDomainRefAttr != nil && typedDomainRefAttr.Name.Space == attr.XBRLDT {
			i := strings.IndexRune(typedDomainRefAttr.Value, '#')
			if i == 0 {
				typedDomainHref += fileName
			}
			typedDomainHref += typedDomainRefAttr.Value
		}
		ret = append(ret, Concept{
			XMLName: xml.Name{
				Space: targetNS,
				Local: nameAttr.Value,
			},
			ID:                idAttr.Value,
			Type:              attr.Xmlns(tlAttrs, typeAttr.Value),
			SubstitutionGroup: substitutionGroup,
			Abstract:          isAbstract,
			Balance:           balance,
			Nillable:          isNillable,
			PeriodType:        periodType,
			TypedDomainHref:   typedDomainHref,
		})
	}
	return ret
}

type Stack []*Concept

// IsEmpty: check if stack is empty
func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

// Push a new value onto the stack
func (s *Stack) Push(concept *Concept) {
	*s = append(*s, concept) // Simply append the new value to the end of the stack
}

// Remove and return top element of stack. Return false if stack is empty.
func (s *Stack) Pop() (*Concept, bool) {
	if s.IsEmpty() {
		return nil, false
	} else {
		index := len(*s) - 1   // Get the index of the top most element.
		element := (*s)[index] // Index into the slice and obtain the element.
		*s = (*s)[:index]      // Remove it from the stack by slicing it off.
		return element, true
	}
}

func (s *Stack) Copy() *Stack {
	dst := make([]*Concept, len(*s))
	copy(dst, *s)
	ret := Stack(dst)
	return &ret
}
