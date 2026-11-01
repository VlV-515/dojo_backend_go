#!/usr/bin/env python3
"""Genera ../cheatsheet-backend-go.pdf a partir de cheatsheet.html.

cheatsheet.html es la fuente: se edita ahí y se regenera. Este script entiende
solo el subconjunto de HTML que usa ese archivo (tarjetas .tarjeta con h2, h3,
p, ul/li, pre, table y avisos .alerta/.ok/.nota) y lo maqueta en 3 columnas
tamaño carta horizontal con reportlab.

Uso (reportlab 4.2.x funciona con el python3 3.9 de macOS):

    python3 -m venv /tmp/venv-cheatsheet
    /tmp/venv-cheatsheet/bin/pip install "reportlab==4.2.5"
    /tmp/venv-cheatsheet/bin/python generar.py
"""

import html
import os
import re
from html.parser import HTMLParser

from reportlab.lib import colors
from reportlab.lib.enums import TA_LEFT
from reportlab.lib.pagesizes import landscape, letter
from reportlab.lib.styles import ParagraphStyle
from reportlab.lib.units import mm
from reportlab.platypus import (BaseDocTemplate, Frame, PageTemplate, Paragraph,
                                Spacer, Table, TableStyle)

AQUI = os.path.dirname(os.path.abspath(__file__))
FUENTE = os.path.join(AQUI, "cheatsheet.html")
SALIDA = os.path.join(AQUI, "..", "cheatsheet-backend-go.pdf")

TINTA = colors.HexColor("#1d2330")
SUAVE = colors.HexColor("#5b6475")
LINEA = colors.HexColor("#d8dce4")
CODE_F = colors.HexColor("#f3f4f7")
AZUL = colors.HexColor("#1f5fae")
COLOR_TARJETA = {
    "": TINTA, "t-go": colors.HexColor("#00788f"), "t-web": colors.HexColor("#6941c6"),
    "t-db": colors.HexColor("#1d7a4a"), "t-docker": colors.HexColor("#1f5fae"),
    "t-arq": colors.HexColor("#b54708"), "t-seg": colors.HexColor("#b42318"),
    "t-dojo": colors.HexColor("#344054"),
}
AVISO = {
    "alerta": (colors.HexColor("#fdecea"), colors.HexColor("#b42318")),
    "ok": (colors.HexColor("#e7f5ee"), colors.HexColor("#1d7a4a")),
    "nota": (colors.HexColor("#fff4e0"), colors.HexColor("#9a5b00")),
}

PAGINA = landscape(letter)
MARGEN = 9 * mm
COLUMNAS = 3
SEPARACION = 4 * mm
ANCHO_COL = (PAGINA[0] - 2 * MARGEN - (COLUMNAS - 1) * SEPARACION) / COLUMNAS
PAD = 4  # relleno interno de la tarjeta (pt)
ANCHO_INTERNO = ANCHO_COL - 2 * PAD

CUERPO = ParagraphStyle("cuerpo", fontName="Helvetica", fontSize=7.2, leading=9.0,
                        textColor=TINTA, alignment=TA_LEFT)
CELDA = ParagraphStyle("celda", parent=CUERPO, fontSize=6.8, leading=8.3)
CABECERA_CELDA = ParagraphStyle("th", parent=CELDA, fontName="Helvetica-Bold", textColor=SUAVE)
H2 = ParagraphStyle("h2", parent=CUERPO, fontName="Helvetica-Bold", fontSize=8.6,
                    leading=10.5, textColor=colors.white)
H3 = ParagraphStyle("h3", parent=CUERPO, fontName="Helvetica-Bold", fontSize=7.5,
                    leading=9.2, textColor=AZUL, spaceBefore=2)
PRE = ParagraphStyle("pre", parent=CUERPO, fontName="Courier", fontSize=6.4, leading=7.7)
LI = ParagraphStyle("li", parent=CUERPO, leftIndent=8, bulletIndent=1)


# --------------------------------------------------------------------------
# Lectura del HTML: arma una lista de tarjetas con sus bloques.
# --------------------------------------------------------------------------

class Lector(HTMLParser):
    """Recorre el HTML y produce tarjetas = {clase, titulo, bloques[]}.

    Cada bloque es (tipo, datos). El texto en línea se convierte al mini
    lenguaje de marcas de reportlab conforme se lee.
    """

    INLINE = {"b", "strong", "i", "em", "code", "small", "span", "br", "a"}

    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.tarjetas = []
        self.titulo = ""
        self.subtitulo = ""
        self.pie = ""
        self.pila = []          # etiquetas abiertas
        self.tarjeta = None
        self.buffer = None      # texto en línea que se está acumulando
        self.en_pre = False
        self.en_titulo = False
        self.tabla = None
        self.fila = None
        self.celda = None
        self.lista = None
        self.aviso = None
        self.zona = None        # "h1", "subtitulo", "pie", None

    # -- helpers ---------------------------------------------------------
    def _abrir_buffer(self):
        self.buffer = []

    def _cerrar_buffer(self):
        texto = "".join(self.buffer or []).strip()
        self.buffer = None
        texto = re.sub(r"^(<br/>)+|(<br/>)+$", "", texto)
        return texto

    def _clases(self, attrs):
        return dict(attrs).get("class", "").split()

    # -- eventos ---------------------------------------------------------
    def handle_starttag(self, tag, attrs):
        clases = self._clases(attrs)
        if tag == "br":  # elemento vacío: no tiene cierre, no va a la pila
            if self.buffer is not None:
                self.buffer.append("<br/>")
            return
        self.pila.append((tag, clases))

        if tag == "h1":
            self.zona = "h1"; self._abrir_buffer(); return
        if tag == "div" and "sub" in clases:
            self.zona = "subtitulo"; self._abrir_buffer(); return
        if tag == "div" and "pie" in clases:
            self.zona = "pie"; self._abrir_buffer(); return
        if tag == "div" and "tarjeta" in clases:
            clase = next((c for c in clases if c.startswith("t-")), "")
            self.tarjeta = {"clase": clase, "titulo": "", "bloques": []}
            return
        if self.tarjeta is None:
            if self.buffer is not None and tag in self.INLINE:
                self._inline_abrir(tag, clases)
            return

        if tag in ("h2", "h3", "p"):
            self.en_titulo = tag == "h2"
            self._abrir_buffer(); return
        if tag == "div" and set(clases) & set(AVISO):
            self.aviso = next(c for c in clases if c in AVISO)
            self._abrir_buffer(); return
        if tag == "pre":
            self.en_pre = True; self._abrir_buffer(); return
        if tag == "ul":
            self.lista = []; return
        if tag == "li":
            self._abrir_buffer(); return
        if tag == "table":
            self.tabla = {"filas": [], "w": dict(attrs).get("data-w")}; return
        if tag == "tr":
            self.fila = []; return
        if tag in ("td", "th"):
            self.celda = {"th": tag == "th", "span": int(dict(attrs).get("colspan", 1))}
            self._abrir_buffer(); return
        if self.buffer is not None and tag in self.INLINE:
            self._inline_abrir(tag, clases)

    def _inline_abrir(self, tag, clases):
        if tag in ("b", "strong"):
            self.buffer.append("<b>")
        elif tag in ("i", "em"):
            self.buffer.append("<i>")
        elif tag == "code" and self.en_titulo:
            self.buffer.append('<font face="Courier-Bold">')
        elif tag == "code" and not self.en_pre:
            self.buffer.append('<font face="Courier" size="6.4" backColor="#f3f4f7">')
        elif tag == "small":
            self.buffer.append('<font size="6.6">')
        elif tag == "span" and "c" in clases:
            self.buffer.append('<font color="#6b7280">')
        elif tag == "span":
            self.buffer.append("<font>")
        elif tag == "br":
            self.buffer.append("<br/>")

    def handle_startendtag(self, tag, attrs):
        if tag == "br" and self.buffer is not None:
            self.buffer.append("<br/>")

    def handle_endtag(self, tag):
        etiqueta, clases = self.pila.pop() if self.pila else (tag, [])

        if tag == "h1" and self.zona == "h1":
            self.titulo = self._cerrar_buffer(); self.zona = None; return
        if tag == "div" and self.zona in ("subtitulo", "pie") and etiqueta == "div" and \
                ("sub" in clases or "pie" in clases):
            texto = self._cerrar_buffer()
            if self.zona == "subtitulo":
                self.subtitulo = texto
            else:
                self.pie = re.sub(r"<[^>]+>", "", texto)
            self.zona = None
            return
        if self.tarjeta is None:
            if self.buffer is not None:
                self._inline_cerrar(tag, clases)
            return

        if tag == "div" and "tarjeta" in clases:
            self.tarjetas.append(self.tarjeta); self.tarjeta = None; return
        if tag == "h2":
            self.en_titulo = False
            self.tarjeta["titulo"] = self._cerrar_buffer(); return
        if tag == "h3":
            self.tarjeta["bloques"].append(("h3", self._cerrar_buffer())); return
        if tag == "p":
            self.tarjeta["bloques"].append(("p", self._cerrar_buffer())); return
        if tag == "div" and self.aviso and set(clases) & set(AVISO):
            self.tarjeta["bloques"].append(("aviso", (self.aviso, self._cerrar_buffer())))
            self.aviso = None
            return
        if tag == "pre":
            self.en_pre = False
            texto = "".join(self.buffer or []).strip("\n")
            self.buffer = None
            self.tarjeta["bloques"].append(("pre", texto))
            return
        if tag == "li":
            self.lista.append(self._cerrar_buffer()); return
        if tag == "ul":
            self.tarjeta["bloques"].append(("ul", self.lista)); self.lista = None; return
        if tag in ("td", "th"):
            self.celda["texto"] = self._cerrar_buffer()
            self.fila.append(self.celda); self.celda = None; return
        if tag == "tr":
            self.tabla["filas"].append(self.fila); self.fila = None; return
        if tag == "table":
            self.tarjeta["bloques"].append(("table", self.tabla)); self.tabla = None; return
        if self.buffer is not None:
            self._inline_cerrar(tag, clases)

    def _inline_cerrar(self, tag, clases):
        if tag in ("b", "strong"):
            self.buffer.append("</b>")
        elif tag in ("i", "em"):
            self.buffer.append("</i>")
        elif tag == "code" and not self.en_pre:
            self.buffer.append("</font>")
        elif tag in ("small", "span"):
            self.buffer.append("</font>")

    def handle_data(self, data):
        if self.buffer is None:
            return
        if self.en_pre:
            self.buffer.append(data)
        else:
            self.buffer.append(html.escape(re.sub(r"\s+", " ", data), quote=False))


# --------------------------------------------------------------------------
# Construcción de flowables.
# --------------------------------------------------------------------------

def pre_a_parrafo(texto):
    """Convierte un bloque <pre> (con <font> de comentarios ya insertados) en un
    Paragraph monoespaciado que respeta saltos y sangría pero puede partir líneas largas."""
    partes = re.split(r"(<font[^>]*>|</font>)", texto)
    salida = []
    for parte in partes:
        if parte.startswith("<font") or parte == "</font>":
            salida.append(parte)
            continue
        parte = html.escape(parte, quote=False)
        lineas = parte.split("\n")
        # La sangría se respeta; los espacios que solo alinean comentarios se
        # reducen a dos para que la línea quepa en la columna sin partirse.
        lineas = [re.sub(r"(?<=\S) {2,}", "&nbsp;&nbsp;", l) for l in lineas]
        lineas = [re.sub(r"^( +)", lambda m: "&nbsp;" * len(m.group(1)), l) for l in lineas]
        salida.append("<br/>".join(lineas))
    return Paragraph("".join(salida), PRE)


def anchos_tabla(tabla, ncols):
    if tabla.get("w"):
        pesos = [float(x) for x in tabla["w"].split(",")]
    else:
        pesos = {1: [1], 2: [42, 58], 3: [20, 40, 40], 4: [22, 28, 18, 32]}.get(ncols, [1] * ncols)
    total = sum(pesos)
    return [ANCHO_INTERNO * p / total for p in pesos]


def tabla_a_flowable(tabla):
    ncols = max(sum(c["span"] for c in f) for f in tabla["filas"])
    datos, estilos = [], []
    for r, fila in enumerate(tabla["filas"]):
        renglon, col = [], 0
        for celda in fila:
            estilo = CABECERA_CELDA if celda["th"] else CELDA
            renglon.append(Paragraph(celda["texto"] or "", estilo))
            if celda["span"] > 1:
                estilos.append(("SPAN", (col, r), (col + celda["span"] - 1, r)))
                renglon.extend([""] * (celda["span"] - 1))
            col += celda["span"]
        renglon.extend([""] * (ncols - len(renglon)))
        datos.append(renglon)
    t = Table(datos, colWidths=anchos_tabla(tabla, ncols))
    t.setStyle(TableStyle([
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (-1, -1), 2), ("RIGHTPADDING", (0, 0), (-1, -1), 2),
        ("TOPPADDING", (0, 0), (-1, -1), 1.2), ("BOTTOMPADDING", (0, 0), (-1, -1), 1.6),
        ("LINEBELOW", (0, 0), (-1, -1), 0.4, LINEA),
    ] + estilos))
    return t


def bloque_a_flowables(tipo, datos):
    if tipo == "h3":
        return [Paragraph(datos, H3)]
    if tipo == "p":
        return [Paragraph(datos, CUERPO)]
    if tipo == "ul":
        return [Paragraph(li, LI, bulletText="•") for li in datos]
    if tipo == "pre":
        return [pre_a_parrafo(datos)]
    if tipo == "table":
        return [tabla_a_flowable(datos)]
    if tipo == "aviso":
        clase, texto = datos
        fondo, borde = AVISO[clase]
        t = Table([[Paragraph(texto, CUERPO)]], colWidths=[ANCHO_INTERNO])
        t.setStyle(TableStyle([
            ("BACKGROUND", (0, 0), (-1, -1), fondo),
            ("LINEBEFORE", (0, 0), (0, -1), 2, borde),
            ("LEFTPADDING", (0, 0), (-1, -1), 4), ("RIGHTPADDING", (0, 0), (-1, -1), 3),
            ("TOPPADDING", (0, 0), (-1, -1), 2), ("BOTTOMPADDING", (0, 0), (-1, -1), 2.5),
        ]))
        return [t]
    raise ValueError(tipo)


def tarjeta_a_flowable(tarjeta):
    """Una tarjeta es una tabla de una columna: fila 0 = título (se repite si la
    tarjeta se parte entre columnas), una fila por bloque."""
    color = COLOR_TARJETA.get(tarjeta["clase"], TINTA)
    filas = [[Paragraph(tarjeta["titulo"], H2)]]
    estilos = []
    for tipo, datos in tarjeta["bloques"]:
        for f in bloque_a_flowables(tipo, datos):
            filas.append([f])
            if tipo == "pre":
                estilos.append(("BACKGROUND", (0, len(filas) - 1), (0, len(filas) - 1), CODE_F))
    t = Table(filas, colWidths=[ANCHO_COL], repeatRows=1)
    t.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (0, 0), color),
        ("BOX", (0, 0), (-1, -1), 0.6, LINEA),
        ("LEFTPADDING", (0, 0), (-1, -1), PAD), ("RIGHTPADDING", (0, 0), (-1, -1), PAD),
        ("TOPPADDING", (0, 0), (-1, -1), 1.2), ("BOTTOMPADDING", (0, 0), (-1, -1), 1.2),
        ("TOPPADDING", (0, 0), (0, 0), 2.5), ("BOTTOMPADDING", (0, 0), (0, 0), 2.5),
        ("BOTTOMPADDING", (0, -1), (-1, -1), 4),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
    ] + estilos))
    return t


# --------------------------------------------------------------------------
# Página.
# --------------------------------------------------------------------------

def main():
    lector = Lector()
    with open(FUENTE, encoding="utf-8") as f:
        lector.feed(f.read())

    alto_cabecera = 15 * mm
    alto_util = PAGINA[1] - 2 * MARGEN - alto_cabecera - 3 * mm
    frames = [Frame(MARGEN + i * (ANCHO_COL + SEPARACION), MARGEN + 3 * mm, ANCHO_COL, alto_util,
                    leftPadding=0, rightPadding=0, topPadding=0, bottomPadding=0, id=f"c{i}")
              for i in range(COLUMNAS)]

    subtitulo = re.sub(r"<[^>]+>", "", lector.subtitulo.replace("<br/>", "  ·  "))

    def decorar(canvas, doc):
        canvas.saveState()
        x0, x1 = MARGEN, PAGINA[0] - MARGEN
        y = PAGINA[1] - MARGEN - 5.5 * mm
        canvas.setFillColor(TINTA)
        canvas.setFont("Helvetica-Bold", 15)
        canvas.drawString(x0, y, re.sub(r"<[^>]+>", "", lector.titulo))
        canvas.setFont("Helvetica", 7.2)
        canvas.setFillColor(SUAVE)
        canvas.drawString(x0, y - 4.2 * mm,
                          "Go · Gin · Swagger · Postgres/GORM · Docker · Hexagonal · Librerías compartidas · CI/CD · Dojo")
        canvas.drawRightString(x1, y + 1, subtitulo[:140])
        canvas.drawRightString(x1, y - 4.2 * mm, f"Página {doc.page}")
        canvas.setStrokeColor(TINTA)
        canvas.setLineWidth(1.2)
        canvas.line(x0, y - 6.5 * mm, x1, y - 6.5 * mm)
        canvas.setFont("Helvetica", 6.2)
        canvas.drawString(x0, MARGEN - 5.5, lector.pie)
        canvas.restoreState()

    doc = BaseDocTemplate(SALIDA, pagesize=PAGINA, leftMargin=MARGEN, rightMargin=MARGEN,
                          topMargin=MARGEN, bottomMargin=MARGEN,
                          title="Cheatsheet Backend Go", author="Dojo Backend Go",
                          subject="Referencia rápida del Dojo Backend Go")
    doc.addPageTemplates([PageTemplate(id="cols", frames=frames, onPage=decorar)])

    historia = []
    for tarjeta in lector.tarjetas:
        historia.append(tarjeta_a_flowable(tarjeta))
        historia.append(Spacer(1, 4))
    doc.build(historia)
    print(f"Listo: {os.path.normpath(SALIDA)} ({len(lector.tarjetas)} tarjetas)")


if __name__ == "__main__":
    main()
