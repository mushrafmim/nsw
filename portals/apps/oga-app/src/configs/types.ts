import type {JsonSchema, UISchemaElement} from "../components/JsonForm";

export interface UIConfig {
  branding: {
    appName: string;
    logoUrl?: string;
    favicon?: string;
  },
  theme?: {
    fontFamily: string;
    borderRadius: string;
  },
  features?: {
    preConsignment: boolean;
    consignmentManagement: boolean;
    reportingDashboard: boolean;
  },
  reviewConfigs: {
    defaultFormId: string;
    forms: {
      reviewType: string;
      reviewDocumentId: string;
      form: {
        schema: JsonSchema;
        uiSchema: UISchemaElement;
      }
    }[]
  }
}