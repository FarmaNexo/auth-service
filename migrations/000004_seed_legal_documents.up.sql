-- migrations/000004_seed_legal_documents.up.sql
-- Seed inicial: tipos de documento + versión 1.0.0 publicada de Términos y Privacidad.
-- También backfilling de user_consents.document_id para consents existentes.

-- ============================================================
-- 1. TIPOS DE DOCUMENTO
-- ============================================================

INSERT INTO auth.legal_document_types (code, name, description, is_required, display_order)
VALUES
    ('terms', 'Términos y Condiciones',
     'Términos legales vinculantes que regulan el uso de la plataforma FarmaNexo.',
     TRUE, 1),
    ('privacy', 'Política de Privacidad',
     'Política de tratamiento de datos personales conforme a la Ley N° 29733 (LPDP).',
     TRUE, 2),
    ('marketing', 'Comunicaciones de Marketing',
     'Consentimiento opcional para recibir comunicaciones promocionales y comerciales.',
     FALSE, 3)
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- 2. VERSIÓN 1.0.0 DE TÉRMINOS Y CONDICIONES
-- ============================================================

INSERT INTO auth.legal_documents (
    document_type_id,
    version,
    locale,
    jurisdiction,
    title,
    summary,
    content_markdown,
    content_hash,
    changelog,
    change_severity,
    status,
    effective_date,
    published_at,
    approved_at,
    metadata
)
SELECT
    t.id,
    '1.0.0',
    'es-PE',
    'PE',
    'Términos y Condiciones de Uso',
    'Documento legal vinculante que regula el acceso y uso de la plataforma FarmaNexo, dirigido a establecimientos farmacéuticos, distribuidoras y profesionales de salud autorizados en el Perú.',
    $LEGALDOC$
**FARMANEXO S.A.C.**

*Plataforma Digital Farmacéutica B2B*

**TÉRMINOS DE USO DE LA PLATAFORMA**

Documento Legal Vinculante — Versión 1.0

# SECCIÓN I: INFORMACIÓN GENERAL

**Artículo 1. Identificación del Titular**

FARMANEXO S.A.C. (en adelante, "Farmanexo" o la "Empresa") es una sociedad anónima cerrada constituida conforme a la Ley N° 26887 — Ley General de Sociedades —, inscrita en los Registros Públicos de la SUNARP, con domicilio legal en la ciudad de Lima, República del Perú.

Farmanexo opera como marketplace digital B2B intermediario para el sector farmacéutico peruano, facilitando la conexión tecnológica entre farmacias independientes, distribuidoras farmacéuticas y profesionales de salud autorizados.

**Artículo 2. Objeto y Alcance**

Los presentes Términos de Uso (en adelante, "los Términos") regulan el acceso y uso de la plataforma digital Farmanexo, disponible en farmanexo.com.pe y sus aplicaciones asociadas (en adelante, "la Plataforma"), por parte de cualquier persona natural o jurídica que se registre o utilice el servicio.

Al crear una cuenta, acceder o usar la Plataforma, el usuario declara haber leído, comprendido y aceptado íntegramente los presentes Términos, así como la Política de Privacidad y demás documentos complementarios publicados en la Plataforma.

# SECCIÓN II: USUARIOS Y REGISTRO

**Artículo 3. Usuarios Autorizados**

Podrán registrarse y utilizar la Plataforma únicamente las siguientes categorías de usuarios:

- Establecimientos farmacéuticos (farmacias, boticas y cadenas farmacéuticas) con autorización vigente expedida por la Dirección General de Medicamentos, Insumos y Drogas (DIGEMID) del Ministerio de Salud (MINSA).

- Empresas distribuidoras de productos farmacéuticos, dispositivos médicos y productos sanitarios, debidamente registradas ante DIGEMID y con RUC activo en SUNAT.

- Profesionales de salud con título habilitante reconocido en el Perú: médicos colegiados (CMP), químicos farmacéuticos (CQFyB), tecnólogos médicos y otros con inscripción vigente en el colegio profesional correspondiente.

- Representantes legales o apoderados de personas jurídicas debidamente acreditados, con poder vigente inscrito en Registros Públicos cuando corresponda.

**Artículo 4. Proceso de Registro**

Para crear una cuenta, el usuario deberá proporcionar información veraz, exacta, actual y completa. Farmanexo se reserva el derecho de verificar la identidad y habilitación del usuario mediante los registros de DIGEMID, SUNAT, SUNARP, CMP, CQFyB u otras entidades oficiales.

El usuario es responsable de mantener actualizada su información de registro. La declaración de datos falsos, inexactos o desactualizados constituye causa suficiente para la suspensión o cancelación inmediata de la cuenta, sin perjuicio de las acciones legales que correspondan conforme al Código Penal peruano (artículo 427 — falsificación de documentos).

**Artículo 5. Credenciales y Seguridad**

El usuario es el único responsable de la confidencialidad de sus credenciales de acceso (usuario y contraseña). Queda expresamente prohibido:

- Ceder, transferir, vender o compartir las credenciales de acceso con terceros.

- Permitir el acceso a la Plataforma a personas no autorizadas mediante su cuenta.

- Utilizar mecanismos automatizados (bots, scripts, crawlers) para acceder a la Plataforma sin autorización expresa de Farmanexo.

Ante cualquier uso no autorizado de la cuenta, el usuario deberá notificar de inmediato a Farmanexo a través de los canales oficiales establecidos en la Plataforma.

# SECCIÓN III: OBLIGACIONES Y CONDUCTA

**Artículo 6. Obligaciones Generales del Usuario**

El usuario se compromete a utilizar la Plataforma exclusivamente para los fines lícitos para los que fue diseñada, conforme a la normativa vigente, los presentes Términos y los estándares de buenas prácticas del sector farmacéutico peruano.

En particular, el usuario se obliga a:

- Cumplir con todas las normas sanitarias aplicables, incluyendo el Reglamento de Establecimientos Farmacéuticos (D.S. 014-2011-SA), la Ley N° 29459 (Ley de Productos Farmacéuticos) y las directivas de DIGEMID.

- No publicar información falsa, engañosa, incorrecta o que pueda inducir a error respecto a medicamentos, dispositivos médicos o productos sanitarios.

- Respetar las normas de farmacovigilancia y notificar Reacciones Adversas a Medicamentos (RAM) a DIGEMID cuando corresponda, conforme al Manual de Buenas Prácticas de Farmacovigilancia.

- No utilizar la Plataforma para comercializar productos sin registro sanitario vigente en DIGEMID, productos falsificados, adulterados, vencidos o sin autorización de comercialización en el Perú.

- No realizar actos de competencia desleal, prácticas monopolísticas o contrarias al Decreto Legislativo N° 1034 (Ley de Represión de Conductas Anticompetitivas).

- Cumplir con las obligaciones tributarias que genere su actividad comercial a través de la Plataforma (emisión de comprobantes de pago conforme a las normas de SUNAT).

**Artículo 7. Conductas Prohibidas**

Queda expresamente prohibido al usuario:

- Acceder o intentar acceder a áreas restringidas de la Plataforma o a los datos de otros usuarios sin autorización.

- Introducir virus, malware, código malicioso u otros elementos que puedan dañar o alterar el funcionamiento de la Plataforma o los sistemas de Farmanexo.

- Realizar ingeniería inversa, descompilar o desensamblar el software de la Plataforma, en contravención del Decreto Legislativo N° 822 (Ley sobre el Derecho de Autor).

- Utilizar la Plataforma para enviar comunicaciones no solicitadas (spam), en contra de la Ley N° 28493 (Ley que Regula el Uso del Correo Electrónico Comercial No Solicitado).

- Suplantar la identidad de otro usuario, de Farmanexo o de cualquier entidad regulatoria.

# SECCIÓN IV: PROPIEDAD INTELECTUAL

**Artículo 8. Titularidad de Derechos**

Todo el contenido de la Plataforma, incluyendo pero no limitado a: marca comercial, logotipo, diseño gráfico, interfaz de usuario, software, código fuente, base de datos de medicamentos, base de datos de establecimientos, contenidos educativos, algoritmos, informes y análisis de datos, es propiedad exclusiva de Farmanexo S.A.C. o de sus licenciantes, y se encuentra protegido por el Decreto Legislativo N° 822 (Ley sobre el Derecho de Autor), el Decreto Legislativo N° 823 (Ley de Propiedad Industrial) y los tratados internacionales suscritos por el Perú en materia de propiedad intelectual.

**Artículo 9. Restricciones de Uso**

Queda expresamente prohibida la reproducción, distribución, comunicación pública, transformación o cualquier otro uso comercial o no comercial del contenido de la Plataforma sin la autorización previa, expresa y por escrito de Farmanexo S.A.C. La infracción a estas disposiciones habilitará a Farmanexo a iniciar las acciones civiles y penales correspondientes conforme a la normativa vigente.

# SECCIÓN V: RESPONSABILIDAD

**Artículo 10. Naturaleza del Servicio y Limitación de Responsabilidad**

Farmanexo actúa exclusivamente como intermediario tecnológico y no es parte en los acuerdos comerciales celebrados entre los usuarios a través de la Plataforma. En consecuencia, y en la máxima medida permitida por la legislación peruana, Farmanexo no asume responsabilidad por:

- La calidad, eficacia, seguridad, origen lícito o condiciones de almacenamiento de los productos farmacéuticos gestionados a través de la Plataforma.

- El incumplimiento de las obligaciones contractuales entre farmacias, distribuidoras u otros usuarios.

- Las decisiones clínicas, terapéuticas o de dispensación tomadas con base en información publicada en la Plataforma.

- Interrupciones del servicio ocasionadas por mantenimiento programado, fuerza mayor, fallas de terceros proveedores de infraestructura o ataques informáticos.

- La veracidad, exactitud o vigencia de la información proporcionada por los usuarios en sus perfiles y publicaciones.

**Artículo 11. Responsabilidad del Usuario frente a Terceros**

El usuario asumirá plena responsabilidad por los daños y perjuicios que ocasione a Farmanexo, a otros usuarios o a terceros como consecuencia del uso indebido de la Plataforma, el incumplimiento de los presentes Términos o la violación de la normativa vigente, incluyendo la responsabilidad civil extracontractual regulada en el artículo 1969 del Código Civil peruano.

# SECCIÓN VI: CUMPLIMIENTO REGULATORIO

**Artículo 12. Marco Normativo Aplicable**

El uso de la Plataforma no exime al usuario de cumplir con todas las normas legales, sanitarias, tributarias, comerciales y de cualquier otra índole aplicables en el territorio peruano. El marco normativo principal aplicable incluye:

**[ Ley N° 26887 — Ley General de Sociedades ] [ Ley N° 29459 — Ley de Productos Farmacéuticos ]**

**[ Ley N° 29571 — Código de Protección al Consumidor ] [ Ley N° 29733 — Protección de Datos Personales ]**

**[ D.S. 014-2011-SA — Reglamento de EE.FF. ] [ D.Leg. 822 — Derecho de Autor ]**

**[ D.Leg. 823 — Propiedad Industrial ] [ D.Leg. 1034 — Libre Competencia ]**

**[ Ley N° 28493 — Correo Electrónico No Solicitado ] [ Normas DIGEMID / MINSA vigentes ]**

Farmanexo cooperará con las autoridades competentes (MINSA, DIGEMID, INDECOPI, SUNAT, Ministerio Público) ante cualquier requerimiento legal o investigación relacionada con el uso de la Plataforma por parte de sus usuarios.

**Artículo 13. Protección al Consumidor**

En el marco del Código de Protección y Defensa del Consumidor (Ley N° 29571), Farmanexo garantiza a los usuarios el acceso a información veraz, suficiente y oportuna sobre las características del servicio. Los usuarios que actúen en calidad de consumidores finales podrán ejercer sus derechos ante el INDECOPI conforme a la legislación vigente.

# SECCIÓN VII: MODIFICACIONES Y TERMINACIÓN

**Artículo 14. Modificaciones a los Términos**

Farmanexo se reserva el derecho de modificar, actualizar o completar los presentes Términos en cualquier momento. Las modificaciones sustanciales serán notificadas al usuario con una anticipación mínima de quince (15) días calendario, mediante el correo electrónico registrado en la cuenta y/o aviso destacado en la Plataforma.

El uso continuado de la Plataforma tras la entrada en vigencia de los nuevos Términos implicará la aceptación plena de los mismos. En caso de desacuerdo, el usuario deberá comunicarlo a Farmanexo y proceder a dar de baja su cuenta.

**Artículo 15. Suspensión y Terminación**

Farmanexo podrá suspender o cancelar el acceso de un usuario a la Plataforma, con o sin previo aviso, en los siguientes supuestos:

- Incumplimiento de los presentes Términos o de cualquier política aplicable.

- Verificación de información falsa en el proceso de registro.

- Uso de la Plataforma para actividades ilegales o contrarias a la normativa sanitaria.

- Resolución judicial o requerimiento de autoridad competente.

- Inactividad prolongada de la cuenta por un período superior a doce (12) meses.

La cancelación de la cuenta no exime al usuario de las obligaciones asumidas durante el período de uso activo.

# SECCIÓN VIII: DISPOSICIONES FINALES

**Artículo 16. Jurisdicción, Ley Aplicable y Solución de Controversias**

Los presentes Términos se rigen e interpretan de conformidad con la legislación de la República del Perú. Las controversias que surjan en relación con la interpretación, validez, ejecución o incumplimiento de los presentes Términos se someterán, en primera instancia, a negociación directa entre las partes.

De no llegarse a un acuerdo en el plazo de treinta (30) días calendario desde el inicio de la negociación, las partes se someterán a la jurisdicción exclusiva de los Juzgados y Tribunales de Lima Cercado, renunciando expresamente a cualquier otro fuero que pudiera corresponderles por razón de domicilio o de otra índole.

**Artículo 17. Idioma y Versión Oficial**

Los presentes Términos se encuentran redactados en idioma español. En caso de discrepancia entre la versión en español y cualquier traducción, prevalecerá la versión en español.

**Artículo 18. Divisibilidad**

Si cualquier disposición de los presentes Términos fuera declarada nula, inválida o inaplicable por un tribunal competente, dicha disposición se tendrá por no puesta, manteniéndose la plena vigencia del resto de los Términos.

**Artículo 19. Datos de Contacto**

Para consultas, reclamos o ejercicio de derechos relacionados con los presentes Términos:

- Correo electrónico: legal@farmanexo.com.pe

- Portal web: farmanexo.com.pe/contacto

- Atención al usuario: lunes a viernes, 9:00 a.m. – 6:00 p.m. (hora de Lima)

*Farmanexo S.A.C. · Lima, Perú · farmanexo.com.pe*

*Versión 1.0 — Enero 2025*

*Marco legal: Ley N° 26887 · Ley N° 29571 · Ley N° 29733 · D.Leg. 822 · D.S. 014-2011-SA · Directivas DIGEMID/MINSA*
$LEGALDOC$,
    'f380ab3f33e6c550d98d29a32d4af9b1688b1aae3c8eeac1625820e49cfbf4f4',
    'Versión inicial publicada — Enero 2025.',
    'major',
    'published',
    '2025-01-01',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '{"published_by": "system_seed", "source": "Scrum/Tyc_Pyp/terminos_uso_farmanexo.docx"}'::jsonb
FROM auth.legal_document_types t
WHERE t.code = 'terms'
ON CONFLICT (document_type_id, version, locale) DO NOTHING;

-- ============================================================
-- 3. VERSIÓN 1.0.0 DE POLÍTICA DE PRIVACIDAD
-- ============================================================

INSERT INTO auth.legal_documents (
    document_type_id,
    version,
    locale,
    jurisdiction,
    title,
    summary,
    content_markdown,
    content_hash,
    changelog,
    change_severity,
    status,
    effective_date,
    published_at,
    approved_at,
    metadata
)
SELECT
    t.id,
    '1.0.0',
    'es-PE',
    'PE',
    'Política de Privacidad y Tratamiento de Datos Personales',
    'Política de tratamiento de datos personales en cumplimiento de la Ley N° 29733 (LPDP) y su reglamento. Detalla finalidades, transferencias, derechos ARCO y medidas de seguridad.',
    $LEGALDOC$
**FARMANEXO S.A.C.**

*Plataforma Digital Farmacéutica B2B*

**POLÍTICA DE PRIVACIDAD Y PROTECCIÓN DE DATOS PERSONALES**

Documento Legal Vinculante — Versión 1.0

# SECCIÓN I: INFORMACIÓN GENERAL

**Artículo 1. Identificación del Responsable del Tratamiento**

FARMANEXO S.A.C. (en adelante, "Farmanexo" o el "Responsable"), con domicilio en Lima, República del Perú, inscrita en los Registros Públicos de la SUNARP conforme a la Ley N° 26887, es responsable del tratamiento de los datos personales de los usuarios de su plataforma digital.

Para ejercer sus derechos o realizar consultas sobre privacidad, el titular puede contactar a Farmanexo a través de: privacidad@farmanexo.com.pe o mediante los canales oficiales disponibles en farmanexo.com.pe.

**Artículo 2. Objeto y Ámbito de Aplicación**

La presente Política de Privacidad y Protección de Datos Personales (en adelante, "la Política") tiene por objeto informar al titular de datos personales sobre la forma en que Farmanexo recopila, usa, trata, almacena, comparte y protege sus datos personales, en cumplimiento de la Ley N° 29733 — Ley de Protección de Datos Personales — y su Reglamento aprobado por D.S. N° 003-2013-JUS, así como las disposiciones complementarias emitidas por la Autoridad Nacional de Protección de Datos Personales del Ministerio de Justicia y Derechos Humanos.

La Política aplica a todos los datos personales recopilados a través de la Plataforma Farmanexo (farmanexo.com.pe), sus aplicaciones móviles, correo electrónico y cualquier otro canal digital oficial de la empresa.

# SECCIÓN II: DATOS PERSONALES RECOPILADOS

**Artículo 3. Categorías de Datos Personales**

Farmanexo recopila las siguientes categorías de datos personales según el tipo de usuario:

**a) Datos de identificación personal:**

- Nombres y apellidos completos.

- Número de documento de identidad (DNI, Carné de Extranjería).

- Correo electrónico personal o corporativo.

- Número de teléfono fijo o celular.

- Fotografía de perfil (opcional).

**b) Datos profesionales y habilitación:**

- Número de colegiatura profesional (CMP, CQFyB u otros).

- Especialidad o profesión declarada.

- Número de autorización o licencia DIGEMID del establecimiento farmacéutico.

- RUC de la persona jurídica o del profesional independiente.

**c) Datos del establecimiento o empresa:**

- Razón social o denominación comercial.

- Dirección fiscal y domicilio del establecimiento.

- Código de farmacia o botica autorizada ante DIGEMID.

**d) Datos de uso y navegación:**

- Historial de búsquedas y consultas en la Plataforma.

- Registros de acceso (logs): dirección IP, fecha, hora, dispositivo utilizado.

- Preferencias de uso y configuración de cuenta.

- Datos de interacción con el chatbot o asistente virtual FarmaBot.

**e) Datos sensibles de salud (cuando aplique):**

En el marco de la Ley N° 29733 (artículo 13), los datos relativos a la salud de las personas son considerados datos sensibles y reciben protección reforzada. Farmanexo puede recopilar datos de salud cuando:

- El profesional de salud los ingresa voluntariamente para fines de consulta clínica o farmacovigilancia.

- Se realicen notificaciones de Reacciones Adversas a Medicamentos (RAM) conforme a las directivas de DIGEMID.

# SECCIÓN III: FINALIDADES Y BASE LEGAL DEL TRATAMIENTO

**Artículo 4. Finalidades del Tratamiento**

Farmanexo trata los datos personales para las siguientes finalidades, conforme al principio de finalidad establecido en el artículo 6 de la Ley N° 29733:

# SECCIÓN IV: DERECHOS DEL TITULAR (DERECHOS ARCO)

**Artículo 5. Derechos Reconocidos**

Conforme a los artículos 19 al 27 de la Ley N° 29733 y los artículos 24 al 35 de su Reglamento (D.S. N° 003-2013-JUS), el titular de los datos personales tiene los siguientes derechos:

**Artículo 6. Procedimiento para el Ejercicio de Derechos ARCO**

Para ejercer cualquiera de los derechos reconocidos en el artículo anterior, el titular deberá remitir una solicitud escrita a privacidad@farmanexo.com.pe, adjuntando:

- Copia de su documento de identidad (DNI o Carné de Extranjería).

- Descripción clara y precisa del derecho que desea ejercer.

- Documentación de sustento, cuando corresponda (por ejemplo, para rectificación).

Farmanexo atenderá la solicitud en un plazo máximo de veinte (20) días hábiles contados desde la recepción, conforme al artículo 24 del Reglamento. En casos de especial complejidad, dicho plazo podrá prorrogarse por diez (10) días hábiles adicionales, previa notificación al titular.

# SECCIÓN V: CONSERVACIÓN, TRANSFERENCIA Y SEGURIDAD

**Artículo 7. Conservación de Datos**

Los datos personales serán conservados durante el tiempo necesario para cumplir con la finalidad para la que fueron recopilados y, adicionalmente, durante los plazos exigidos por la normativa peruana vigente:

- Datos de carácter tributario: mínimo cinco (5) años, conforme al Código Tributario (D.S. N° 133-2013-EF).

- Datos relacionados con obligaciones sanitarias y farmacovigilancia: conforme a las directivas de DIGEMID y MINSA vigentes.

- Datos de registro de cuenta activa: durante la vigencia de la relación contractual.

- Datos de cuenta cancelada: hasta un máximo de dos (2) años para fines de trazabilidad y seguridad, salvo mayor plazo legal.

Transcurridos dichos plazos, los datos serán eliminados de forma segura o anonimizados de manera irreversible.

**Artículo 8. Transferencia de Datos a Terceros**

Farmanexo no vende, cede ni transfiere datos personales a terceros con fines comerciales propios de dichos terceros. Las transferencias de datos se limitan a los siguientes supuestos:

- Proveedores tecnológicos de infraestructura y servicios cloud (hosting, bases de datos, seguridad informática), bajo contrato de confidencialidad y garantías de seguridad equivalentes a las de la presente Política.

- Autoridades sanitarias y de supervisión (DIGEMID, MINSA, INDECOPI, SUNAT, Ministerio Público), cuando sean requeridos por ley o en el marco de investigaciones oficiales.

- Socios comerciales farmacéuticos registrados en la Plataforma, únicamente con el consentimiento expreso del titular y en la medida estrictamente necesaria para la prestación del servicio.

Las transferencias internacionales de datos, en caso de realizarse, se efectuarán únicamente hacia países que cuenten con un nivel adecuado de protección o bajo garantías contractuales apropiadas, conforme a los artículos 15 y 16 de la Ley N° 29733.

**Artículo 9. Medidas de Seguridad**

Farmanexo implementa medidas técnicas, organizativas y jurídicas para garantizar la seguridad de los datos personales y prevenir su pérdida, acceso no autorizado, divulgación, alteración o destrucción, entre las que se incluyen:

- Cifrado de datos en tránsito mediante protocolos HTTPS/TLS.

- Control de acceso basado en roles (RBAC) con principio de mínimo privilegio.

- Registro y auditoría de accesos (logs) con monitoreo continuo.

- Procedimientos de gestión de incidentes de seguridad y notificación a la Autoridad Nacional conforme al artículo 29 de la Ley N° 29733.

- Políticas internas de seguridad de la información alineadas con estándares internacionales (ISO/IEC 27001).

- Pseudonimización y anonimización de datos para usos analíticos agregados.

# SECCIÓN VI: CONSENTIMIENTO Y COMUNICACIONES

**Artículo 10. Consentimiento Informado**

El tratamiento de datos personales por parte de Farmanexo se basa en el consentimiento libre, previo, expreso, informado e inequívoco del titular, obtenido en el momento del registro, conforme al artículo 13 de la Ley N° 29733.

El usuario puede revocar su consentimiento en cualquier momento, sin efecto retroactivo, enviando una solicitud a privacidad@farmanexo.com.pe. La revocación del consentimiento no afectará la legalidad del tratamiento efectuado con anterioridad.

**Artículo 11. Comunicaciones Comerciales y Opt-Out**

El envío de comunicaciones comerciales, boletines y novedades de Farmanexo requiere el consentimiento expreso del usuario mediante el marcado voluntario del casillero habilitado en el formulario de registro.

El usuario puede darse de baja en cualquier momento de las comunicaciones comerciales utilizando el enlace de cancelación de suscripción incluido en cada correo, o enviando una solicitud a privacidad@farmanexo.com.pe. Las comunicaciones de carácter estrictamente operativo (alertas DIGEMID, cambios en los Términos, notificaciones de seguridad) no son susceptibles de opt-out mientras la cuenta esté activa.

**Artículo 12. Cookies y Tecnologías de Rastreo**

Farmanexo utiliza cookies y tecnologías similares para garantizar el funcionamiento técnico de la Plataforma y mejorar la experiencia del usuario. Las cookies se clasifican en:

- Cookies estrictamente necesarias: imprescindibles para el funcionamiento de la Plataforma; no requieren consentimiento.

- Cookies analíticas y de rendimiento: permiten analizar el uso de la Plataforma de forma agregada y anonimizada; requieren consentimiento.

- Cookies de personalización: permiten recordar preferencias del usuario; requieren consentimiento.

El usuario puede gestionar sus preferencias de cookies desde la configuración de su navegador o mediante el panel de privacidad disponible en su cuenta. La desactivación de cookies necesarias puede afectar el funcionamiento de la Plataforma.

# SECCIÓN VII: DISPOSICIONES ESPECIALES

**Artículo 13. Menores de Edad**

Farmanexo es una plataforma orientada exclusivamente al sector profesional y empresarial farmacéutico (B2B). No recopila ni trata datos personales de menores de dieciocho (18) años de forma deliberada. Si Farmanexo toma conocimiento de que ha recopilado datos de un menor de edad, procederá de inmediato a la cancelación de la cuenta y eliminación de sus datos, notificándolo a la Autoridad Nacional de Protección de Datos Personales si correspondiera.

**Artículo 14. Datos de Terceros Ingresados por el Usuario**

Si el usuario ingresa a la Plataforma datos personales de terceros (pacientes, empleados, representantes), declara contar con el consentimiento o la habilitación legal necesaria para ello, asumiendo plena responsabilidad por dicho tratamiento. Farmanexo tratará estos datos conforme a la presente Política.

**Artículo 15. Incidentes de Seguridad**

Ante un incidente de seguridad que afecte datos personales y que genere un riesgo para los derechos y libertades de los titulares, Farmanexo procederá a notificar a la Autoridad Nacional de Protección de Datos Personales y, cuando corresponda, a los titulares afectados, en los plazos y formas previstos por la Ley N° 29733 y su Reglamento.

# SECCIÓN VIII: ACTUALIZACIÓN Y CONTACTO

**Artículo 16. Actualización de la Política**

Farmanexo se reserva el derecho de actualizar o modificar la presente Política en cualquier momento para reflejar cambios normativos, tecnológicos o en los servicios prestados. Las modificaciones serán notificadas al titular con al menos quince (15) días de anticipación a su entrada en vigencia, mediante correo electrónico y/o aviso destacado en la Plataforma.

El uso continuado de la Plataforma tras la entrada en vigencia de la nueva versión de la Política implicará la aceptación de los cambios.

**Artículo 17. Contacto y Delegado de Protección de Datos**

Para consultas, solicitudes de ejercicio de derechos ARCO, reclamos o cualquier asunto relacionado con el tratamiento de datos personales:

- Correo electrónico de privacidad: privacidad@farmanexo.com.pe

- Portal web: farmanexo.com.pe/privacidad

- Dirección postal: Lima, República del Perú (consultar dirección actualizada en farmanexo.com.pe)

- Horario de atención: lunes a viernes, 9:00 a.m. – 6:00 p.m. (hora de Lima, Perú)

Farmanexo designará oportunamente un Delegado de Protección de Datos (DPD) o responsable interno de privacidad, cuyo contacto será publicado en la Plataforma.

**Artículo 18. Autoridad de Control**

Si el titular considera que sus derechos en materia de protección de datos personales no han sido debidamente atendidos por Farmanexo, tiene derecho a presentar una reclamación ante la Autoridad Nacional de Protección de Datos Personales del Ministerio de Justicia y Derechos Humanos del Perú:

- Portal web: www.minjus.gob.pe

- Mesa de partes: Jr. Scipión Llona 350, Miraflores, Lima, Perú

*Farmanexo S.A.C. · Lima, Perú · farmanexo.com.pe*

*Versión 1.0 — Enero 2025*

*Marco legal: Ley N° 29733 · D.S. N° 003-2013-JUS · Resolución Directoral N° 019-2013-JUS/DGPDP · Ley N° 29571*
$LEGALDOC$,
    'ba000a1f85c777a9c617ddd2d3a698506fbdd22912f8df88065cb6fc85366e08',
    'Versión inicial publicada — Enero 2025.',
    'major',
    'published',
    '2025-01-01',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    '{"published_by": "system_seed", "source": "Scrum/Tyc_Pyp/politica_privacidad_farmanexo.docx"}'::jsonb
FROM auth.legal_document_types t
WHERE t.code = 'privacy'
ON CONFLICT (document_type_id, version, locale) DO NOTHING;

-- ============================================================
-- 4. BACKFILL: user_consents existentes apuntan al documento equivalente
-- ============================================================
-- Mapea consent_type legacy → code del legal_document_types y matchea por document_version.

UPDATE auth.user_consents uc
SET document_id = ld.id,
    content_hash_at_acceptance = ld.content_hash
FROM auth.legal_documents ld
JOIN auth.legal_document_types lt ON lt.id = ld.document_type_id
WHERE uc.document_id IS NULL
  AND ld.locale = 'es-PE'
  AND (
        (uc.consent_type = 'terms_of_service'         AND lt.code = 'terms')
     OR (uc.consent_type = 'privacy_policy'           AND lt.code = 'privacy')
     OR (uc.consent_type = 'marketing_communications' AND lt.code = 'marketing')
      )
  -- Match más reciente: si el consent.document_version coincide con la versión del documento,
  -- se mapea exacto; si no coincide (documento todavía no seedeado), se mapea a la versión
  -- publicada actual como mejor aproximación (el hash siempre queda registrado para auditoría).
  AND ld.version = COALESCE(
        (SELECT version FROM auth.legal_documents
         WHERE document_type_id = ld.document_type_id
           AND locale = 'es-PE'
           AND version = uc.document_version
         LIMIT 1),
        ld.version
      );
