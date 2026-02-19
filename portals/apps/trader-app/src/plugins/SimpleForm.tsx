import {JsonForm, type JsonSchema, type UISchemaElement, useJsonForm} from "../components/JsonForm";
import {sendTaskCommand} from "../services/task.ts";
import {uploadFile} from "../services/upload";
import {useLocation, useNavigate, useParams} from "react-router-dom";
import {useState} from "react";
import {Button} from "@radix-ui/themes";


export interface TaskFormData {
  title: string
  schema: JsonSchema
  uiSchema: UISchemaElement
  formData: Record<string, unknown>
}

export type SimpleFormConfig = {
  traderFormInfo: TaskFormData
  ogaReviewForm?: TaskFormData
  submissionResponseForm?: TaskFormData
}

function TraderForm(props: { formInfo: TaskFormData, pluginState: string }) {
  const {consignmentId, preConsignmentId, taskId} = useParams<{
    consignmentId?: string
    preConsignmentId?: string
    taskId?: string
  }>()
  const location = useLocation()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  const READ_ONLY_STATES = ['OGA_REVIEWED', 'SUBMITTED', 'OGA_ACKNOWLEDGED'];
  const isReadOnly = READ_ONLY_STATES.includes(props.pluginState);

  const isPreConsignment = location.pathname.includes('/pre-consignments/')
  const workflowId = preConsignmentId || consignmentId

  const replaceFilesWithKeys = async (value: unknown): Promise<unknown> => {
    if (value instanceof File) {
      const metadata = await uploadFile(value)
      return metadata.key
    }

    if (Array.isArray(value)) {
      return await Promise.all(value.map(replaceFilesWithKeys))
    }

    if (value && typeof value === 'object') {
      const entries = await Promise.all(
        Object.entries(value as Record<string, unknown>).map(async ([key, nested]) => [
          key,
          await replaceFilesWithKeys(nested),
        ] as const)
      )
      return Object.fromEntries(entries)
    }

    return value
  }

  const handleSubmit = async (data: unknown) => {
    if (!workflowId || !taskId) {
      setError('Workflow ID or Task ID is missing.')
      return
    }

    try {
      setError(null)

      // Send form submission - data now contains file keys (strings) instead of File objects
      const preparedData = await replaceFilesWithKeys(data) as Record<string, unknown>

      const response = await sendTaskCommand({
        command: 'SUBMISSION',
        taskId,
        workflowId,
        data: preparedData,
      })

      if (response.success) {
        // Navigate back to appropriate workflow list
        navigate(isPreConsignment ? '/pre-consignments' : `/consignments/${workflowId}`)
      } else {
        setError(response.error?.message || 'Failed to submit form.')
      }
    } catch (err) {
      console.error('Error submitting form:', err)
      setError('Failed to submit form. Please try again.')
    }
  }

  const form = useJsonForm({
    schema: props.formInfo.schema,
    data: props.formInfo.formData,
    onSubmit: handleSubmit,
  })

  const showAutoFillButton = import.meta.env.VITE_SHOW_AUTOFILL_BUTTON === 'true'

  return (
    <>
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <h1 className="text-2xl font-bold text-gray-800">{props.formInfo.title}</h1>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6">
        <form onSubmit={form.handleSubmit} noValidate>
          <JsonForm
            schema={props.formInfo.schema}
            uiSchema={props.formInfo.uiSchema}
            values={form.values}
            errors={form.errors}
            touched={form.touched}
            setValue={form.setValue}
            setTouched={form.setTouched}
            readOnly={isReadOnly}
          />
          {!isReadOnly && (
            <div className={`mt-4 flex gap-3 ${showAutoFillButton ? 'justify-between' : ''}`}>
              {showAutoFillButton && (
                <Button
                  type="button"
                  variant="soft"
                  color="purple"
                  size={"3"}
                  className={"flex-1!"}
                  onClick={form.autoFillForm}
                  disabled={form.isSubmitting}
                >
                  Demo - Auto Fill
                </Button>
              )}
              <Button
                type="submit"
                disabled={form.isSubmitting}
                className={'flex-1!'}
                size={"3"}
              >
                {form.isSubmitting ? 'Submitting...' : 'Submit Form'}
              </Button>
            </div>
          )}
        </form>
      </div>

      {error && (
        <div className="bg-red-100 text-red-700 rounded-lg p-4 mt-4">
          <p>{error}</p>
        </div>
      )}
    </>
  )
}

function SubmissionResponseForm(props: { formInfo: TaskFormData }) {
  const form = useJsonForm({
    schema: props.formInfo.schema,
    data: props.formInfo.formData,
    onSubmit: () => {
    },
  })

  return (
    <div className="mt-6 border-l-4 border-emerald-500 rounded-r-lg overflow-hidden shadow-sm">
      <div className="bg-emerald-50 px-6 py-4 flex items-center gap-3">
        <span className="inline-flex items-center justify-center w-7 h-7 rounded-full bg-emerald-100 text-emerald-700 shrink-0">
          <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd"/>
          </svg>
        </span>
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-emerald-600 mb-0.5">Submission Response</p>
          <h2 className="text-lg font-bold text-emerald-900 leading-tight">{props.formInfo.title}</h2>
        </div>
      </div>

      <div className="bg-white border-t border-emerald-100 p-6">
        <JsonForm
          schema={props.formInfo.schema}
          uiSchema={props.formInfo.uiSchema}
          values={form.values}
          errors={form.errors}
          touched={form.touched}
          setValue={form.setValue}
          setTouched={form.setTouched}
          readOnly={true}
        />
      </div>
    </div>
  )
}

function OgaReviewForm(props: { formInfo: TaskFormData }) {
  const form = useJsonForm({
    schema: props.formInfo.schema,
    data: props.formInfo.formData,
    onSubmit: () => {
    },
  })

  return (
    <div className="mt-6 rounded-lg overflow-hidden shadow-sm border border-indigo-200">
      <div className="bg-indigo-700 px-6 py-4 flex items-center gap-3">
        <span className="inline-flex items-center justify-center w-7 h-7 rounded-full bg-indigo-600 text-indigo-100 shrink-0">
          <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
            <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/>
            <path fillRule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd"/>
          </svg>
        </span>
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-indigo-300 mb-0.5">OGA Review</p>
          <h2 className="text-lg font-bold text-white leading-tight">{props.formInfo.title}</h2>
        </div>
      </div>

      <div className="bg-indigo-50 p-6">
        <JsonForm
          schema={props.formInfo.schema}
          uiSchema={props.formInfo.uiSchema}
          values={form.values}
          errors={form.errors}
          touched={form.touched}
          setValue={form.setValue}
          setTouched={form.setTouched}
          readOnly={true}
        />
      </div>
    </div>
  )
}

export default function SimpleForm(props: { configs: SimpleFormConfig, pluginState: string }) {
  return (
    <div>
      <TraderForm formInfo={props.configs.traderFormInfo} pluginState={props.pluginState}/>

      {props.configs.submissionResponseForm && (
        <SubmissionResponseForm formInfo={props.configs.submissionResponseForm}/>
      )}

      {props.configs.ogaReviewForm && (
        <OgaReviewForm formInfo={props.configs.ogaReviewForm}/>
      )}
    </div>
  )
}
