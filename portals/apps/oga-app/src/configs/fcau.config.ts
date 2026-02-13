import type {UIConfig} from "./types.ts";

export const config: UIConfig = {
  branding: {
    appName: "Food Control Administration Unit (FCAU)",
    logoUrl: "",
    favicon: ""
  },
  reviewConfigs: {
    defaultFormId: "moh:fcau:health_cert:002",
    forms: [
      {
        reviewType: "consignment",
        reviewDocumentId: "moh:fcau:health_cert:001",
        form: {
          schema: {
            "type": "object",
            "required": ["decision", "foodSafetyClearance"],
            "properties": {
              "decision": {
                "type": "string",
                "title": "Decision",
                "oneOf": [
                  {"const": "APPROVED", "title": "Approved"},
                  {"const": "REJECTED", "title": "Rejected"}
                ]
              },
              "foodSafetyClearance": {
                "type": "string",
                "title": "Food Safety Compliance Status",
                "oneOf": [
                  {"const": "COMPLIANT", "title": "Compliant - Approved for Export"},
                  {"const": "MINOR_NON_COMPLIANCE", "title": "Minor Non-Compliance"},
                  {"const": "MAJOR_NON_COMPLIANCE", "title": "Major Non-Compliance"}
                ]
              },
              "labReportReference": {
                "type": "string",
                "title": "Laboratory Report Reference No"
              },
              "remarks": {
                "type": "string",
                "title": "FCAU Remarks"
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
                "scope": "#/properties/foodSafetyClearance"
              },
              {
                "type": "Control",
                "scope": "#/properties/labReportReference"
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
        reviewDocumentId: "moh:fcau:health_cert:002",
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
    ],
  }
}