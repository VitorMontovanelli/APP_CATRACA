export namespace main {
	
	export class AccessLog {
	    id: number;
	    user_id?: number;
	    // Go type: time
	    timestamp: any;
	    resultado: string;
	    motivo: string;
	
	    static createFrom(source: any = {}) {
	        return new AccessLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.resultado = source["resultado"];
	        this.motivo = source["motivo"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AcessoEstatistica {
	    label: string;
	    liberado: number;
	    negado: number;
	
	    static createFrom(source: any = {}) {
	        return new AcessoEstatistica(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.liberado = source["liberado"];
	        this.negado = source["negado"];
	    }
	}
	export class AlunoPorPlano {
	    student_id: number;
	    nome: string;
	    cpf: string;
	    status: string;
	    start_date: string;
	
	    static createFrom(source: any = {}) {
	        return new AlunoPorPlano(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.student_id = source["student_id"];
	        this.nome = source["nome"];
	        this.cpf = source["cpf"];
	        this.status = source["status"];
	        this.start_date = source["start_date"];
	    }
	}
	export class AuditLog {
	    id: number;
	    actor_id: number;
	    action: string;
	    target_id?: number;
	    details: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new AuditLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.actor_id = source["actor_id"];
	        this.action = source["action"];
	        this.target_id = source["target_id"];
	        this.details = source["details"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BackupConfig {
	    token: string;
	    chat_id: string;
	    auto_backup: boolean;
	    last_backup: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.chat_id = source["chat_id"];
	        this.auto_backup = source["auto_backup"];
	        this.last_backup = source["last_backup"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class CobrancaFatura {
	    id: number;
	    amount_cents: number;
	    status: string;
	    due_date: string;
	    paid_at?: string;
	    comprovante?: string;
	    dias_vencido: number;
	
	    static createFrom(source: any = {}) {
	        return new CobrancaFatura(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.amount_cents = source["amount_cents"];
	        this.status = source["status"];
	        this.due_date = source["due_date"];
	        this.paid_at = source["paid_at"];
	        this.comprovante = source["comprovante"];
	        this.dias_vencido = source["dias_vencido"];
	    }
	}
	export class CobrancaAluno {
	    student_id: number;
	    nome: string;
	    cpf: string;
	    plano_nome: string;
	    plano_preco: number;
	    status_plano: string;
	    due_day: number;
	    faturas: CobrancaFatura[];
	
	    static createFrom(source: any = {}) {
	        return new CobrancaAluno(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.student_id = source["student_id"];
	        this.nome = source["nome"];
	        this.cpf = source["cpf"];
	        this.plano_nome = source["plano_nome"];
	        this.plano_preco = source["plano_preco"];
	        this.status_plano = source["status_plano"];
	        this.due_day = source["due_day"];
	        this.faturas = this.convertValues(source["faturas"], CobrancaFatura);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class InadimplenteReport {
	    student_id: number;
	    nome: string;
	    cpf: string;
	    telefone?: string;
	    email?: string;
	    plano_nome: string;
	    plano_preco: number;
	    status_plano: string;
	    invoice_id: number;
	    valor_devido: number;
	    data_vencimento: string;
	    dias_vencido: number;
	    ultimo_pagamento?: string;
	    total_em_aberto: number;
	
	    static createFrom(source: any = {}) {
	        return new InadimplenteReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.student_id = source["student_id"];
	        this.nome = source["nome"];
	        this.cpf = source["cpf"];
	        this.telefone = source["telefone"];
	        this.email = source["email"];
	        this.plano_nome = source["plano_nome"];
	        this.plano_preco = source["plano_preco"];
	        this.status_plano = source["status_plano"];
	        this.invoice_id = source["invoice_id"];
	        this.valor_devido = source["valor_devido"];
	        this.data_vencimento = source["data_vencimento"];
	        this.dias_vencido = source["dias_vencido"];
	        this.ultimo_pagamento = source["ultimo_pagamento"];
	        this.total_em_aberto = source["total_em_aberto"];
	    }
	}
	export class Invoice {
	    id: number;
	    student_plan_id: number;
	    student_id: number;
	    plan_id: number;
	    amount_cents: number;
	    status: string;
	    payment_method_id?: number;
	    due_date: string;
	    // Go type: time
	    paid_at?: any;
	    paid_amount_cents?: number;
	    gateway_transaction_id?: string;
	    pix_qr_code?: string;
	    pix_br_code?: string;
	    comprovante?: string;
	    notes?: string;
	    // Go type: time
	    created_at: any;
	    student_name: string;
	    plan_name: string;
	    payment_method_name: string;
	
	    static createFrom(source: any = {}) {
	        return new Invoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.student_plan_id = source["student_plan_id"];
	        this.student_id = source["student_id"];
	        this.plan_id = source["plan_id"];
	        this.amount_cents = source["amount_cents"];
	        this.status = source["status"];
	        this.payment_method_id = source["payment_method_id"];
	        this.due_date = source["due_date"];
	        this.paid_at = this.convertValues(source["paid_at"], null);
	        this.paid_amount_cents = source["paid_amount_cents"];
	        this.gateway_transaction_id = source["gateway_transaction_id"];
	        this.pix_qr_code = source["pix_qr_code"];
	        this.pix_br_code = source["pix_br_code"];
	        this.comprovante = source["comprovante"];
	        this.notes = source["notes"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.student_name = source["student_name"];
	        this.plan_name = source["plan_name"];
	        this.payment_method_name = source["payment_method_name"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanoContador {
	    nome: string;
	    qtd: number;
	
	    static createFrom(source: any = {}) {
	        return new PlanoContador(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nome = source["nome"];
	        this.qtd = source["qtd"];
	    }
	}
	export class MetricasHome {
	    total_usuarios: number;
	    total_admins: number;
	    total_alunos: number;
	    total_faturado_cents: number;
	    total_atraso_cents: number;
	    total_pendente_cents: number;
	    planos_distribuicao: PlanoContador[];
	
	    static createFrom(source: any = {}) {
	        return new MetricasHome(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_usuarios = source["total_usuarios"];
	        this.total_admins = source["total_admins"];
	        this.total_alunos = source["total_alunos"];
	        this.total_faturado_cents = source["total_faturado_cents"];
	        this.total_atraso_cents = source["total_atraso_cents"];
	        this.total_pendente_cents = source["total_pendente_cents"];
	        this.planos_distribuicao = this.convertValues(source["planos_distribuicao"], PlanoContador);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaymentGateway {
	    id: number;
	    name: string;
	    type: string;
	    enabled: boolean;
	    config?: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new PaymentGateway(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.config = source["config"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaymentMethod {
	    id: number;
	    name: string;
	    type: string;
	    enabled: boolean;
	    fee_percent: number;
	    fee_fixed_cents: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new PaymentMethod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.fee_percent = source["fee_percent"];
	        this.fee_fixed_cents = source["fee_fixed_cents"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Plan {
	    id: number;
	    name: string;
	    description: string;
	    duration_days: number;
	    price_cents: number;
	    grace_period_days: number;
	    active: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Plan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.duration_days = source["duration_days"];
	        this.price_cents = source["price_cents"];
	        this.grace_period_days = source["grace_period_days"];
	        this.active = source["active"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ResultadoAcesso {
	    nome: string;
	    liberado: boolean;
	    mensagem: string;
	
	    static createFrom(source: any = {}) {
	        return new ResultadoAcesso(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nome = source["nome"];
	        this.liberado = source["liberado"];
	        this.mensagem = source["mensagem"];
	    }
	}
	export class Student {
	    id: number;
	    nome: string;
	    cpf: string;
	    data_nascimento?: string;
	    telefone?: string;
	    email?: string;
	    forma_pagamento_id?: number;
	    // Go type: time
	    data_entrada: any;
	    observacao?: string;
	    ativo: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Student(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.cpf = source["cpf"];
	        this.data_nascimento = source["data_nascimento"];
	        this.telefone = source["telefone"];
	        this.email = source["email"];
	        this.forma_pagamento_id = source["forma_pagamento_id"];
	        this.data_entrada = this.convertValues(source["data_entrada"], null);
	        this.observacao = source["observacao"];
	        this.ativo = source["ativo"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StudentComPlano {
	    id: number;
	    nome: string;
	    cpf: string;
	    data_nascimento?: string;
	    telefone?: string;
	    email?: string;
	    forma_pagamento_id?: number;
	    // Go type: time
	    data_entrada: any;
	    observacao?: string;
	    ativo: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    plano_nome?: string;
	    plano_preco?: number;
	    plano_status?: string;
	    student_plan_id?: number;
	    payment_method?: string;
	    due_day?: number;
	    ultima_fatura?: string;
	    fatura_valor?: number;
	    fatura_vencimento?: string;
	
	    static createFrom(source: any = {}) {
	        return new StudentComPlano(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.cpf = source["cpf"];
	        this.data_nascimento = source["data_nascimento"];
	        this.telefone = source["telefone"];
	        this.email = source["email"];
	        this.forma_pagamento_id = source["forma_pagamento_id"];
	        this.data_entrada = this.convertValues(source["data_entrada"], null);
	        this.observacao = source["observacao"];
	        this.ativo = source["ativo"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.plano_nome = source["plano_nome"];
	        this.plano_preco = source["plano_preco"];
	        this.plano_status = source["plano_status"];
	        this.student_plan_id = source["student_plan_id"];
	        this.payment_method = source["payment_method"];
	        this.due_day = source["due_day"];
	        this.ultima_fatura = source["ultima_fatura"];
	        this.fatura_valor = source["fatura_valor"];
	        this.fatura_vencimento = source["fatura_vencimento"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StudentPlan {
	    id: number;
	    student_id: number;
	    plan_id: number;
	    status: string;
	    payment_method_id?: number;
	    due_day: number;
	    // Go type: time
	    start_date: any;
	    // Go type: time
	    end_date?: any;
	    // Go type: time
	    cancelled_at?: any;
	    notes?: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at?: any;
	    student_name: string;
	    plan_name: string;
	    plan_price_cents: number;
	
	    static createFrom(source: any = {}) {
	        return new StudentPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.student_id = source["student_id"];
	        this.plan_id = source["plan_id"];
	        this.status = source["status"];
	        this.payment_method_id = source["payment_method_id"];
	        this.due_day = source["due_day"];
	        this.start_date = this.convertValues(source["start_date"], null);
	        this.end_date = this.convertValues(source["end_date"], null);
	        this.cancelled_at = this.convertValues(source["cancelled_at"], null);
	        this.notes = source["notes"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.student_name = source["student_name"];
	        this.plan_name = source["plan_name"];
	        this.plan_price_cents = source["plan_price_cents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class User {
	    id: number;
	    name: string;
	    email: string;
	    cargo: string;
	    registrador_id?: number;
	    ativo: boolean;
	    foto?: string;
	    permissoes: string;
	    // Go type: time
	    criado_em: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.cargo = source["cargo"];
	        this.registrador_id = source["registrador_id"];
	        this.ativo = source["ativo"];
	        this.foto = source["foto"];
	        this.permissoes = source["permissoes"];
	        this.criado_em = this.convertValues(source["criado_em"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

