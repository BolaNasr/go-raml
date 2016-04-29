// Copyright 2014 DoAT. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation and/or
//    other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED “AS IS” WITHOUT ANY WARRANTIES WHATSOEVER.
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
// THE IMPLIED WARRANTIES OF NON INFRINGEMENT, MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE HEREBY DISCLAIMED. IN NO EVENT SHALL DoAT OR CONTRIBUTORS
// BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// // THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
// EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// The views and conclusions contained in the software and documentation are those of
// the authors and should not be interpreted as representing official policies,
// either expressed or implied, of DoAT.

// This package contains the parser, validator and types that implement the
// RAML specification, as documented here:
// http://raml.org/spec.html
package raml

// This file contains all of the RAML types.

import "strings"

// TODO: We don't support !include of non-text files. RAML supports including
//       of many file types.

// "Any" type, for our convenience
type Any interface{}

// For extra clarity
type HTTPCode int      // e.g. 200
type HTTPHeader string // e.g. Content-Length

// Headers used in Methods and other types
type Header NamedParameter

// All documentation of the API is of this format.
type Documentation struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

// Some method verbs expect the resource to be sent as a request body.
// For example, to create a resource, the request must include the details of
// the resource to create.
// Resources CAN have alternate representations. For example, an API might
// support both JSON and XML representations.
type Body struct {
	mediaType string `yaml:"mediaType"`
	// TODO: Fill this during the post-processing phase

	// The structure of a request or response body MAY be further specified
	// by the schema property under the appropriate media type.
	// The schema key CANNOT be specified if a body's media type is
	// application/x-www-form-urlencoded or multipart/form-data.
	// All parsers of RAML MUST be able to interpret JSON Schema [JSON_SCHEMA]
	// and XML Schema [XML_SCHEMA].
	// Alternatively, the value of the schema field MAY be the name of a schema
	// specified in the root-level schemas property
	Schema string `yaml:"schema"`

	// Brief description
	Description string `yaml:"description"`

	// Example attribute to generate example invocations
	Example string `yaml:"example"`

	// Web forms REQUIRE special encoding and custom declaration.
	// If the API's media type is either application/x-www-form-urlencoded or
	// multipart/form-data, the formParameters property MUST specify the
	// name-value pairs that the API is expecting.
	// The formParameters property is a map in which the key is the name of
	// the web form parameter, and the value is itself a map the specifies
	// the web form parameter's attributes
	FormParameters map[string]NamedParameter `yaml:"formParameters"`
	// TODO: This doesn't make sense in response bodies.. separate types for
	// request and response body?

	Headers map[HTTPHeader]Header `yaml:"headers"`
}

// Container of Body types, necessary because of technical reasons.
type Bodies struct {

	// Instead of using a simple map[HTTPHeader]Body for the body
	// property of the Response and Method, we use the Bodies struct. Why?
	// Because some RAML APIs don't use the MIMEType part, instead relying
	// on the mediaType property in the APIDefinition.
	// So, you might see:
	//
	// responses:
	//   200:
	//     body:
	//       example: "some_example" : "123"
	//
	// and also:
	//
	// responses:
	//   200:
	//     body:
	//       application/json:
	//         example: |
	//           {
	//             "some_example" : "123"
	//           }

	// As in the Body type.
	DefaultSchema string `yaml:"schema"`

	// As in the Body type.
	DefaultDescription string `yaml:"description"`

	// As in the Body type.
	DefaultExample string `yaml:"example"`

	// As in the Body type.
	DefaultFormParameters map[string]NamedParameter `yaml:"formParameters"`

	// TODO: Is this ever used? I think I put it here by mistake.
	//Headers               map[HTTPHeader]Header     `yaml:"headers"`

	// Resources CAN have alternate representations. For example, an API
	// might support both JSON and XML representations. This is the map
	// between MIME-type and the body definition related to it.
	ForMIMEType map[string]Body `yaml:",regexp:.*"`

	// TODO: For APIs without a priori knowledge of the response types for
	// their responses, "*/*" MAY be used to indicate that responses that do
	// not matching other defined data types MUST be accepted. Processing
	// applications MUST match the most descriptive media type first if
	// "*/*" is used.
	ApplicationJson *BodiesProperty `yaml:"application/json"`

	Type string `yaml:"type"`
}

// Resource methods MAY have one or more responses.
type Response struct {

	// HTTP status code of the response
	HTTPCode HTTPCode
	// TODO: Fill this during the post-processing phase

	// Clarifies why the response was emitted. Response descriptions are
	// particularly useful for describing error conditions.
	Description string

	// An API's methods may support custom header values in responses
	Headers map[HTTPHeader]Header `yaml:"headers"`

	// TODO: API's may include the the placeholder token {?} in a header name
	// to indicate that any number of headers that conform to the specified
	// format can be sent in responses. This is particularly useful for
	// APIs that allow HTTP headers that conform to some naming convention
	// to send arbitrary, custom data.

	// Each response MAY contain a body property. Responses that can return
	// more than one response code MAY therefore have multiple bodies defined.
	Bodies Bodies `yaml:"body"`
}

// A ResourceType/Trait/SecurityScheme choice contains the name of a
// ResourceType/Trait/SecurityScheme as well as the parameters used to create
// an instance of it.
// Parameters MUST be of type string.
type DefinitionParameters map[string]interface{}
type DefinitionChoice struct {
	Name string

	// The definitions of resource types and traits MAY contain parameters,
	// whose values MUST be specified when applying the resource type or trait,
	// UNLESS the parameter corresponds to a reserved parameter name, in which
	// case its value is provided by the processing application.
	// Same goes for security schemes.
	Parameters DefinitionParameters
}

// Unmarshal a node which MIGHT be a simple string or a
// map[string]DefinitionParameters
func (dc *DefinitionChoice) UnmarshalYAML(unmarshaler func(interface{}) error) error {

	simpleDefinition := new(string)
	parameterizedDefinition := make(map[string]DefinitionParameters)

	var err error

	// Unmarshal into a string
	if err = unmarshaler(simpleDefinition); err == nil {
		dc.Name = *simpleDefinition
		dc.Parameters = nil
	} else if err = unmarshaler(parameterizedDefinition); err == nil {
		// Didn't work? Now unmarshal into a map
		for choice, params := range parameterizedDefinition {
			dc.Name = choice
			dc.Parameters = params
		}
	}

	// Still didn't work? Panic

	return err
}

// A trait-like structure to a security scheme mechanism so as to extend
// the mechanism, such as specifying response codes, HTTP headers or custom
// documentation.
type SecuritySchemeMethod struct {
	//Bodies          Bodies                    `yaml:"body"`
	Headers         map[HTTPHeader]Header     `yaml:"headers"`
	QueryParameters map[string]NamedParameter `yaml:"queryParameters"`
	Responses       map[HTTPCode]Response     `yaml:"responses"`
}

// Most REST APIs have one or more mechanisms to secure data access, identify
// requests, and determine access level and data visibility.
type SecurityScheme struct {
	Name string
	// TODO: Fill this during the post-processing phase

	// Briefly describes the security scheme
	Description string `yaml:"description"`

	// The type attribute MAY be used to convey information about
	// authentication flows and mechanisms to processing applications
	// such as Documentation Generators and Client generators.
	Type string `yaml:"type"`
	// TODO: Verify that it is of the values accepted: "OAuth 1.0",
	// "OAuth 2.0", "Basic Authentication", "Digest Authentication",
	// "x-{other}"

	// The describedBy attribute MAY be used to apply a trait-like structure
	// to a security scheme mechanism so as to extend the mechanism, such as
	// specifying response codes, HTTP headers or custom documentation.
	// This extension allows API designers to describe security schemes.
	// As a best practice, even for standard security schemes, API designers
	// SHOULD describe the security schemes' required artifacts, such as
	// headers, URI parameters, and so on.
	// Including the security schemes' description completes an API's documentation.
	DescribedBy SecuritySchemeMethod `yaml:"describedBy"`

	// The settings attribute MAY be used to provide security schema-specific
	// information. Depending on the value of the type parameter, its attributes
	// can vary.
	Settings map[string]Any `yaml:"settings"`
	// TODO: Verify OAuth 1.0, 2.0 settings
	// TODO: Add to documentaiotn

	// If the scheme's type is x-other, API designers can use the properties
	// in this mapping to provide extra information to clients that understand
	// the x-other type.
	Other map[string]string
}

// Methods are operations that are performed on a resource
type Method struct {
	Name string

	// Briefly describes what the method does to the resource
	Description string `yaml:"description"`

	DisplayName string `yaml:"displayName"`

	// Applying a securityScheme definition to a method overrides whichever
	// securityScheme has been defined at the root level. To indicate that
	// the method is protected using a specific security scheme, the method
	// MUST be defined by using the securedBy attribute
	// Custom parameters can be provided to the security scheme.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`
	// TODO: To indicate that the method may be called without applying any
	// securityScheme, the method may be annotated with the null securityScheme.

	// Object whose property names are the query parameter names
	// and whose values describe the values.
	Headers map[HTTPHeader]Header `yaml:"headers"`
	// TODO: Examples for headers are REQUIRED.
	// TODO: If the header name contains the placeholder token {*}, processing
	// applications MUST allow requests to send any number of headers that
	// conform to the format specified, with {*} replaced by 0 or more valid
	// header characters, and offer a way for implementations to add an
	// arbitrary number of such headers. This is particularly useful for APIs
	// that allow HTTP headers that conform to custom naming conventions to
	// send arbitrary, custom data.

	// A RESTful API method can be reached HTTP, HTTPS, or both.
	// A method can override an API's protocols value for that single method
	// by setting a different value for the fields.
	Protocols []string `yaml:"protocols"`

	// The queryParameters property is a map in which the key is the query
	// parameter's name, and the value is itself a map specifying the query
	//  parameter's attributes
	QueryParameters map[string]NamedParameter `yaml:"queryParameters"`

	// Some method verbs expect the resource to be sent as a request body.
	// A method's body is defined in the body property as a hashmap, in which
	// the key MUST be a valid media type.
	Bodies Bodies `yaml:"body"`
	// TODO: Check - how does the mediaType play play here? What it do?

	// Resource methods MAY have one or more responses. Responses MAY be
	// described using the description property, and MAY include example
	// attributes or schema properties.
	// Responses MUST be a map of one or more HTTP status codes, where each
	// status code itself is a map that describes that status code.
	Responses map[HTTPCode]Response `yaml:"responses"`

	// Methods may specify one or more traits from which they inherit using the
	// is property
	Is []DefinitionChoice `yaml:"is"`
	// TODO: Add support for inline traits?
}

// A resource is the conceptual mapping to an entity or set of entities.
type Resource struct {

	// Resources are identified by their relative URI, which MUST begin with
	// a slash (/).
	URI string

	// A resource defined as a child property of another resource is called a
	// nested resource, and its property's key is its URI relative to its
	// parent resource's URI. If this is not nil, then this resource is a
	// child resource.
	Parent *Resource

	// A friendly name to the resource
	DisplayName string `yaml:"displayName"`

	// Briefly describes the resource
	Description string `yaml:"description"`

	// A securityScheme may also be applied to a resource by using the
	// securedBy key, which is equivalent to applying the securityScheme to
	// all methods of this Resource.
	// Custom parameters can be provided to the security scheme.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`
	// TODO: To indicate that the method may be called without applying any
	// securityScheme, the method may be annotated with the null securityScheme.

	// Template URIs containing URI parameters can be used to define a
	// resource's relative URI when it contains variable elements.
	// The values matched by URI parameters cannot contain slash (/) characters
	URIParameters map[string]NamedParameter `yaml:"uriParameters"`

	// TODO: If a URI parameter in a resource's relative URI is not explicitly
	// described in a uriParameters property for that resource, it MUST still
	// be treated as a URI parameter with defaults as specified in the Named
	// Parameters section of this specification. Its type is "string", it is
	// required, and its displayName is its name (i.e. without the surrounding
	// curly brackets [{] and [}]). In the example below, the top-level
	// resource has two URI parameters, "folderId" and "fileId

	// TOOD: A special uriParameter, mediaTypeExtension, is a reserved
	// parameter. It may be specified explicitly in a uriParameters property
	// or not specified explicitly, but its meaning is reserved: it is used
	// by a client to specify that the body of the request or response be of
	// the associated media type. By convention, a value of .json is
	// equivalent to an Accept header of application/json and .xml is
	// equivalent to an Accept header of text/xml.

	// Resources may specify the resource type from which they inherit using
	// the type property. The resource type may be defined inline as the value
	// of the type property (directly or via an !include), or the value of
	// the type property may be the name of a resource type defined within
	// the root-level resourceTypes property.
	// NOTE: inline not currently supported.
	Type *DefinitionChoice `yaml:"type"`

	// TODO: Add support for inline ResourceTypes

	// A resource may use the is property to apply the list of traits to all
	// its methods.
	Is []DefinitionChoice `yaml:"is"`
	// TODO: Add support for inline traits?

	// In a RESTful API, methods are operations that are performed on a
	// resource. A method MUST be one of the HTTP methods defined in the
	// HTTP version 1.1 specification [RFC2616] and its extension,
	// RFC5789 [RFC5789].
	Get     *Method `yaml:"get"`
	Patch   *Method `yaml:"patch"`
	Put     *Method `yaml:"put"`
	Head    *Method `yaml:"head"`
	Post    *Method `yaml:"post"`
	Delete  *Method `yaml:"delete"`
	Options *Method `yaml:"options"`

	// A resource defined as a child property of another resource is called a
	// nested resource, and its property's key is its URI relative to its
	// parent resource's URI.
	Nested map[string]*Resource `yaml:",regexp:/.*"`

	// all methods of this resource
	Methods []*Method `yaml:"-"`
}

// TODO: Resource.GetBaseURIParameter --> includeds APIDefinition BURIParams..
// TODO: Resource.GetAbsoluteURI

// The API Definition describes the basic information of an API, such as its
// title and base URI, and describes how to define common schema references.
type APIDefinition struct {

	// RAML 0.8
	RAMLVersion string `yaml:"raml_version"`

	// The title property is a short plain text description of the RESTful API.
	// The title property's value SHOULD be suitable for use as a title for the
	// contained user documentation
	Title string `yaml:"title"`

	// If RAML API definition is targeted to a specific API version, it should
	// be noted here
	Version string `yaml:"version"`

	// A RESTful API's resources are defined relative to the API's base URI.
	// If the baseUri value is a Level 1 Template URI, the following reserved
	// base URI parameters are available for replacement:
	//
	// version - The content of the version field.
	BaseUri string `yaml:"baseUri"`
	// TODO: If a URI template variable in the base URI is not explicitly
	// described in a baseUriParameters property, and is not specified in a
	// resource-level baseUriParameters property, it MUST still be treated as
	// a base URI parameter with defaults as specified in the Named Parameters
	//  section of this specification. Its type is "string", it is required,
	// and its displayName is its name (i.e. without the surrounding curly
	// brackets [{] and [}]).

	// A resource or a method can override a base URI template's values.
	// This is useful to restrict or change the default or parameter selection
	// in the base URI. The baseUriParameters property MAY be used to override
	// any or all parameters defined at the root level baseUriParameters
	// property, as well as base URI parameters not specified at the root level.
	// In a resource structure of resources and nested resources with their
	// methods, the most specific baseUriParameter fully overrides any
	// baseUriParameter definition made before
	BaseUriParameters map[string]NamedParameter `yaml:"baseUriParameters"`
	// TODO: Generate these also from the baseUri

	// Level 1 URI custom parameters, which are useful in a variety of scenario.
	// URI parameters can be further defined by using the uriParameters
	// property. The use of uriParameters is OPTIONAL. The uriParameters
	// property MUST be a map in which each key MUST be the name of the URI
	// parameter as defined in the baseUri property. The uriParameters CANNOT
	// contain a key named version because it is a reserved URI parameter name.
	UriParameters map[string]NamedParameter `yaml:"uriParameters"`

	// A RESTful API can be reached HTTP, HTTPS, or both
	Protocols []string `yaml:"protocols"`

	// The media types returned by API responses, and expected from API
	// requests that accept a body, MAY be defaulted by specifying the
	// mediaType property.
	// The property's value MAY be a single string with a valid media type:
	//
	// One of the following YAML media types:
	// * text/yaml
	// * text/x-yaml
	// * application/yaml
	// * application/x-yaml*
	//
	// Any type from the list of IANA MIME Media Types,
	// http://www.iana.org/assignments/media-types
	// A custom type that conforms to the regular expression:
	// * "application\/[A-Za-z.-0-1]*+?(json|xml)"
	MediaType string `yaml:"mediaType"`

	// The schemas property specifies collections of schemas that could be
	// used anywhere in the API definition.
	// The value of the schemas property is an array of maps; in each map,
	// the keys are the schema name, and the values are schema definitions:
	// []map[SchemaName]SchemaString
	Schemas []map[string]string
	// TODO: Flatten the arrays of maps here.

	// The securitySchemes property MUST be used to specify an API's security
	// mechanisms, including the required settings and the authentication
	// methods that the API supports.
	// []map[SchemeName]SecurityScheme
	SecuritySchemes []map[string]SecurityScheme `yaml:"securitySchemes"`
	// TODO: Flatten the arrays of maps here.

	// To apply a securityScheme definition to every method in an API, the
	// API MAY be defined using the securedBy attribute. This specifies that
	// all methods in the API are protected using that security scheme.
	// Custom parameters can be provided to the security scheme.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`

	// The API definition can include a variety of documents that serve as a
	// user guides and reference documentation for the API. Such documents can
	// clarify how the API works or provide business context.
	// All the sections are in the order in which the documentation is declared.
	Documentation []Documentation `yaml:"documentation"`

	// To apply a trait definition to a method, so that the method inherits the
	// trait's characteristics, the method MUST be defined by using the is
	// attribute. The value of the is attribute MUST be an array of any number
	// of elements, each of which MUST be a) one or more trait keys (names)
	// included in the traits declaration, or b) one or more trait definition
	// maps.
	Traits []map[string]Trait `yaml:"traits"`

	// The resourceTypes and traits properties are declared at the API
	// definition's root level with the resourceTypes and traits property keys,
	// respectively. The value of each of these properties is an array of maps;
	// in each map, the keys are resourceType or trait names, and the values
	// are resourceType or trait definitions, respectively.
	ResourceTypes []map[string]ResourceType `yaml:"resourceTypes"`

	// Resources are identified by their relative URI, which MUST begin with a
	// slash (/). A resource defined as a root-level property is called a
	// top-level resource. Its property's key is the resource's URI relative
	// to the baseUri. A resource defined as a child property of another
	// resource is called a nested resource, and its property's key is its
	// URI relative to its parent resource's URI.
	Resources map[string]Resource `yaml:",regexp:/.*"`

	Types map[string]Type `yaml:"types"`
}

// This function receives a path, splits it and traverses the resource
// tree to find the appropriate resource
func (r *APIDefinition) GetResource(path string) *Resource {
	return nil
}

// Property defines a Type property
type Property struct {
	Name     string
	Type     string      `yaml:"type"`
	Required bool        `yaml:"required"`
	Enum     interface{} `yaml:"enum"`

	// string
	Pattern   *string
	MinLength *int
	MaxLength *int

	// number
	Minimum    *float64
	Maximum    *float64
	MultipleOf *float64
	//Format *string

	// array
	MinItems    *int
	MaxItems    *int
	UniqueItems bool
}

// ToProperty creates a property from an interface
// we use `interface{}` as property type to support syntactic sugar & shortcut
func ToProperty(name string, p interface{}) Property {
	// convert number(int/float) to float
	toFloat64 := func(number interface{}) float64 {
		switch v := number.(type) {
		case int:
			return float64(v)
		case float64:
			return v
		default:
			return v.(float64)
		}
	}
	// convert from map of interface to property
	mapToProperty := func(val map[interface{}]interface{}) Property {
		var p Property
		p.Required = true
		for k, v := range val {
			switch k {
			case "type":
				p.Type = v.(string)
			case "required":
				p.Required = v.(bool)
			case "enum":
				p.Enum = v
			case "minLength":
				p.MinLength = new(int)
				*p.MinLength = v.(int)
			case "maxLength":
				p.MaxLength = new(int)
				*p.MaxLength = v.(int)
			case "pattern":
				p.Pattern = new(string)
				*p.Pattern = v.(string)
			case "minimum":
				p.Minimum = new(float64)
				*p.Minimum = toFloat64(v)
			case "maximum":
				p.Maximum = new(float64)
				*p.Maximum = toFloat64(v)
			case "multipleOf":
				p.MultipleOf = new(float64)
				*p.MultipleOf = toFloat64(v)
			case "minItems":
				p.MinItems = new(int)
				*p.MinItems = v.(int)
			case "maxItems":
				p.MaxItems = new(int)
				*p.MaxItems = v.(int)
			case "uniqueItems":
				p.UniqueItems = v.(bool)
			}
		}
		return p
	}

	prop := Property{Required: true}
	switch p.(type) {
	case string:
		prop.Type = p.(string)
	case map[interface{}]interface{}:
		prop = mapToProperty(p.(map[interface{}]interface{}))
	case Property:
		prop = p.(Property)
	}

	if prop.Type == "" { // if has no type, we set it as string
		prop.Type = "string"
	}

	prop.Name = name

	// if has "?" suffix, remove the "?" and set required=false
	if strings.HasSuffix(prop.Name, "?") {
		prop.Required = false
		prop.Name = prop.Name[:len(prop.Name)-1]
	}
	return prop

}

type Type struct {
	// Alias for the equivalent "type" property,
	// for compatibility with RAML 0.8.
	// Deprecated - API definitions should use the "type" property,
	// as the "schema" alias for that property name may be removed in a future RAML version.
	// The "type" property allows for XML and JSON schemas.
	Schema interface{} `yaml:"schema"`

	// An example of an instance of this type.
	// This can be used, e.g., by documentation generators to generate sample values for an object of this type.
	// Cannot be present if the examples property is present.
	Example interface{} `yaml:"example"`

	// An object containing named examples of instances of this type.
	Examples map[string]interface{} `yaml:"examples"`

	// A base type which the current type extends,
	// or more generally a type expression.
	Type interface{} `yaml:"type"`

	// An alternate, human-friendly name for the type
	DisplayName string `yaml:"displayName"`

	// A longer, human-friendly description of the type.
	Description string `yaml:"description"`

	// The properties that instances of this type may or must have.
	// we use `interface{}` as property type to support syntactic sugar & shortcut
	Properties map[string]interface{} `yaml:"properties"`

	// JSON schema style syntax for declaring maps
	AdditionalProperties string `yaml:"additionalProperties"`

	// Enum type
	Enum interface{} `yaml:"enum"`

	// Type property name to be used as a discriminator or boolean
	Discriminator string `yaml:"discriminator"`

	// Validators
	MinItems    int  `yaml:"minItems"`
	MaxItems    int  `yaml:"maxItems"`
	UniqueItems bool `yaml:"uniqueItems"`
}

// IsMap checks if a type is a Map type as defined in http://docs.raml.org/specs/1.0/#raml-10-spec-types
// map types could be written in these forms:
// - a `[]` property
// - a regex within `[]` property. example : [a-zA-Z]
// - additionalProperties fied in Type
// - patternProperties filed in Type TODO
// Type's type must be `object`
func (t Type) IsMap() bool {
	// check if this map type written using `[]`
	squareBracketPropCheck := func() bool {
		if len(t.Properties) != 1 {
			return false
		}
		for k := range t.Properties {
			if strings.HasPrefix(k, "[") && strings.HasSuffix(k, "]") {
				return true
			}
		}
		return false
	}

	if squareBracketPropCheck() {
		return true
	}

	if t.AdditionalProperties != "" {
		return true
	}
	return false
}

// IsArray checks if this type is an Array
// see specs at http://docs.raml.org/specs/1.0/#raml-10-spec-array-types
func (t Type) IsArray() bool {
	return strings.HasSuffix(t.Type.(string), "[]")
}

// IsEnum type check if this type is an enum
// http://docs.raml.org/specs/1.0/#raml-10-spec-enums
func (t Type) IsEnum() bool {
	return t.Enum != nil
}

// IsUnion checks if a type is Union type
// see http://docs.raml.org/specs/1.0/#raml-10-spec-union-types
func (t Type) IsUnion() bool {
	return strings.Index(t.Type.(string), "|") > 0
}

// BodiesProperty defines a Body's property
type BodiesProperty struct {
	// we use `interface{}` as property type to support syntactic sugar & shortcut
	Properties map[string]interface{} `yaml:"properties"`

	Type string
}
