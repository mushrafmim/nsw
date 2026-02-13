import type {UIConfig} from "./types.ts";

export const config: UIConfig = {
  branding: {
    appName: "National Plant Quarantine Service (NPQS)",
    logoUrl: "",
    favicon: ""
  },
  reviewConfigs: {
    defaultFormId: "moa:npqs:phytosanitary:002",
    forms: [
      {
        reviewType: "consignment",
        reviewDocumentId: "moa:npqs:phytosanitary:001",
        form: {
          schema: {
            "type": "object",
            "required": ["decision", "phytosanitaryClearance"],
            "properties": {
              "decision": {
                "type": "string",
                "title": "Decision",
                "oneOf": [
                  {"const": "APPROVED", "title": "Approved"},
                  {"const": "REJECTED", "title": "Rejected"}
                ]
              },
              "phytosanitaryClearance": {
                "type": "string",
                "title": "Phytosanitary Clearance Status",
                "oneOf": [
                  {"const": "CLEARED", "title": "Cleared for Export"},
                  {"const": "CONDITIONAL", "title": "Cleared with Conditions"},
                  {"const": "REJECTED", "title": "Rejected - Non Compliance"}
                ]
              },
              "inspectionReference": {
                "type": "string",
                "title": "Inspection / Certificate Reference No"
              },
              "remarks": {
                "type": "string",
                "title": "NPQS Remarks"
              }
            }
          },
          uiSchema: {
            "type": "VerticalLayout",
            "elements": [
              {
                "type": "Control",
                "scope": "#/properties/decision"
              },
              {
                "type": "Control",
                "scope": "#/properties/phytosanitaryClearance"
              },
              {
                "type": "Control",
                "scope": "#/properties/inspectionReference"
              },
              {
                "type": "Control",
                "scope": "#/properties/remarks",
                "options": {"multi": true}
              }
            ]
          }
        }
      },
      {
        reviewType: "consignment",
        reviewDocumentId: "moa:npqs:phytosanitary:002",
        form: {
          schema: {
            "type": "object",
            "required": ["decision", "reviewedAt"],
            "properties": {
              "decision": {
                "enum": [
                  "APPROVED",
                  "REJECTED",
                  "NEEDS_MORE_INFO"
                ],
                "type": "string"
              },
              "reviewedAt": {
                "type": "string",
                "format": "date-time"
              },
              "reviewerNotes": {
                "type": "string"
              }
            }
          },
          uiSchema: {
            "type": "VerticalLayout",
            "elements": [
              {
                "type": "Control",
                "scope": "#/properties/decision",
                "options": {"format": "radio"}
              },
              {
                "type": "Control",
                "scope": "#/properties/reviewerNotes",
                "options": {"multi": true}
              },
              {
                "type": "Control",
                "scope": "#/properties/reviewedAt"
              }
            ]
          }
        }
      }
    ]
  }
}