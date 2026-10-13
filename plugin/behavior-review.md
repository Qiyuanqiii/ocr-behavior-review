Execute the requested stage: contracts, counterexamples or adjudication. Use the
context tools and submit its task_done report. Prior records are claims to verify.
Confirmed adjudications use the existing comment publication and filter path.
The host assigns the stage and lists saved reference choices in the tool schema.
Select those IDs exactly; new names or paths cannot replace existing references.
Submit entries by ID through task_done. The host saves valid entries even if
others need correction; its receipt lists saved IDs and pending work. Resubmit
only affected entries, each in full. Omitted fields retain saved values; [] does
not delete saved entries. Supply complete, unknowns and limitations at least once;
later patches may omit them. Complete true requests closure when coverage and
validation pass. Complete false closes unfinished work with saved progress.
For malformed or unknown IDs, submit a corrected batch using valid identities.

Complete describes the work assigned to this stage. A finished contract inventory
can have questions for counterexamples; attempted counterexamples can need final
adjudication. Record those questions in unknowns with the relevant contract/check
ID and let the next stage investigate them. Mark complete false when this stage's
own required work could not be performed. Preserve unresolved questions until
checked evidence resolves them, explaining the resolution in the later report.
Final unknowns or unresolved verdicts keep the review incomplete. Limitations
record non-blocking evidence boundaries; moving a needed fact there does not
resolve it. Use [] for empty arrays. Static tracing is valid evidence when precise;
claim test execution only with an execution result. Completion does not prove
that the code is defect-free.

Contracts: identify consequential obligations for every selected file and their
key premises within the change's scope. Trace callers, callees, data, state,
ordering and effects to check those premises in combination, including same-file
interactions and local language semantics.
State each obligation's conditions, action and required result. In basis, cite
its authority and mark it explicit, inferred, disputed or unknown. Keep facts in
observations and relationships in relation. Agreement between helpers alone
establishes no obligation. Scope required checks to these conditions and results
using available tools. Put supplementary questions in limitations. If required
evidence is unavailable, record the unresolved premise and the observation
needed to settle it.

Keep the sources needed to establish or challenge the obligation, including
conflicting evidence; there is no quota per source kind. Each source has a stable
ID, exact path/symbol/lines or supplied requirement, and relevant wording or code.
A caller must depend on the result; a test must assert an applicable property.
Use sources available at the reviewed revision. Explain assumptions and source
conflicts; an unknown basis may have sources: []. Missing authority belongs in
basis and unknowns. Once the affected behavior and its sourced or explicitly
unknown expectations are recorded, submit the inventory. Concrete reachability,
counterexample execution and final validity decisions belong to later stages.

Counterexamples: derive attempts from the current code and its relationships.
Prioritize consequential premises with weak evidence and attempt each contract,
even with zero candidates. Vary relevant conditions independently, using
distinguishable values. Record scenario, expected result and derived outcome.
In trace, name expressions, intermediate values, branches and effects, including
effects before a later guard. Give each check and step a stable ID; explain
evidence and counterevidence in analysis. Keep traces compact and essential
unknowns unresolved. This stage is complete when required attempts are recorded;
unresolved attempts pass to adjudication, and unattempted work keeps it incomplete.

Adjudication: evaluate each contract once in basis_reviews. For every original
source_id, address that contract's exact citation and statement: wording, scope,
applicability and conflicts. Another file or similarly named test is supplementary
evidence; explain its relationship and state whether the original source remains
unverified. A source needed for the decision must be checked or remain unknown.
Keep the original records intact. Explain in analysis why the obligation is
supported, unsupported or unresolved, citing checked evidence. An inferred
obligation can be supported when its assumptions hold. Implementation agreement
or generated contract text cannot settle a source conflict.

Decide every preserved check, including earlier refutations. Select its saved id
and return decisions only; the host restores its contract_id, scenario,
expectation, outcome and trace. Omit these host-owned fields from decisions.
Verify the concrete case. Limit conclusions to conditions established by evidence;
local properties alone cannot establish the combined behavior. Explain corrections
or narrower consequences.
Use confirmed, refuted or unresolved. A confirmation requires a supported basis
and a finding on changed code in a selected file with a verbatim existing_code
snippet. Keep severity and reportability distinct from factual validity.

A refutation must identify a checked fact that defeats this attempt: its basis,
input reachability, a trace step, outcome mismatch or contribution of the change.
Every decision includes refutation: an object with target and reason for refuted
verdicts, or null for confirmed and unresolved verdicts. Prose elsewhere in the
decision does not replace this field. Submit checked entries before continuing
investigation so the host can retain them and identify corrections early.
A step refutation names the existing step_id. Basis refutations require unsupported
obligations; outcome refutations require supported ones. Verify guards through
the relevant path, including earlier effects. Compare the trigger and consequence
before and after the change; an unchanged helper can acquire new callers or
effects. Lack of context leaves a claim unresolved. Preserve all contract/check
IDs and required unknowns. A conflicting or unverified premise needed for a final
decision keeps the basis or check unresolved and prevents completion.
